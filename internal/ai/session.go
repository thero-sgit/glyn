package ai

import (
	"context"
	"errors"
	"io"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/ai/common"
	co "github.com/thero-sgit/glyn/internal/common"
	"github.com/thero-sgit/glyn/internal/persistence"
)

type Session struct {
	Ctx   context.Context
	cancl context.CancelFunc
	groq  *groq
	Wg    sync.WaitGroup
	agent *agent.Agent

	History []common.Message

	//
	lifecycleMu     sync.Mutex
	sessionStore    *persistence.Store
	msgToDiscChan   chan common.MessageWrapper
	dbWorkerStop    chan struct{}
	dbWorkerDone    chan error
	dbWorkerStarted bool
	activeSessionID string
	closeOnce       sync.Once
	closeErr        error
	errMu           sync.RWMutex
	errChan         chan error
}

func (s *Session) SetNewContext() {
	if s.cancl != nil {
		s.cancl()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.Ctx = ctx
	s.cancl = cancel
}

func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		s.lifecycleMu.Lock()
		defer s.lifecycleMu.Unlock()

		if s.cancl != nil {
			s.cancl()
		}
		s.Wg.Wait()
		s.closeErr = errors.Join(s.stopDBWorker(), s.sessionStore.Close())
	})
	return s.closeErr
}

func (s *Session) NameSession(message string, errChan chan error) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if len(s.History) > 0 || s.activeSessionID != "" {
		return
	}
	s.setErrorChannel(errChan)

	req := StreamRequest{
		Model: "openai/gpt-oss-20b",
		Messages: []common.Message{
			{Role: "system", Content: "Generate a 3-5 word title for this prompt. Return ONLY the title. MUST BE 3 TO 5 WORDS!"},
			{Role: "user", Content: message},
		},
		Stream: true,
	}

	s.Wg.Add(1)

	var name string
	ctx := s.Ctx
	go func() {
		defer s.Wg.Done()

		requestCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		const idleTimeout = 15 * time.Second
		timer := time.AfterFunc(idleTimeout, cancel)
		defer timer.Stop()

		stream, err := s.groq.CreateChatCompletionStream(requestCtx, req)
		if err != nil {
			s.handleError(err, errChan)
			return
		}
		defer stream.Close()

		for {
			select {
			case <-requestCtx.Done():
				return
			default:
			}

			response, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					c, err := s.sessionStore.StartNewSession(ctx, name)
					if err != nil {
						s.handleError(err, errChan)
						return
					}
					s.activeSessionID = c.Id
					s.startDBWorker(ctx, c.Id)
					return
				}
				s.handleError(err, errChan)
				return
			}

			if len(response.Choices) > 0 {
				name += response.Choices[0].Delta.Content
			}

			timer.Reset(idleTimeout)
		}
	}()
}

func (s *Session) NewPrompt(message string, c chan common.Chunk, errChan chan error, eventsChan chan<- common.Event) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	h := append(s.History, common.Message{
		Role:    "user",
		Content: message,
	})

	h = *common.PruneAndFit(&h)

	s.setErrorChannel(errChan)
	now := time.Now().Format(co.TsFormat)

	s.Wg.Add(1)
	ctx := s.Ctx
	go func() {
		defer s.Wg.Done()
		defer close(c)
		defer close(eventsChan)

		requestCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		h, err := s.agent.Run(requestCtx, s.groq.handlePrompt, h, now, c, eventsChan, s.msgToDiscChan)
		if err != nil {
			s.handleError(err, errChan)
			return
		}
		s.History = h
	}()
}

func (s *Session) ReadSessions() ([]persistence.ChatSession, error) {
	return s.sessionStore.ListSessions(s.Ctx)
}

func (s *Session) ContinueSession(sessionId string) ([]common.MessageWrapper, error) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	if s.cancl != nil {
		s.cancl()
	}

	s.Wg.Wait()
	s.SetNewContext()
	s.setErrorChannel(nil)

	oldSessionID := s.activeSessionID
	stopErr := s.stopDBWorker()

	h, err := s.sessionStore.ListMessagesInSession(s.Ctx, sessionId)
	if err != nil {
		if oldSessionID != "" {
			s.startDBWorker(s.Ctx, oldSessionID)
		}
		return nil, errors.Join(err, stopErr)
	}

	go s.unwrapMessages(h)
	s.activeSessionID = sessionId
	s.startDBWorker(s.Ctx, sessionId)
	return h, stopErr
}

func (s *Session) unwrapMessages(messages []common.MessageWrapper) {
	var history []common.Message
	for i := range messages {
		history = append(history, messages[i].Msg)
	}
	s.History = history
}

func (s *Session) writeToDbWorker(notifyCtx context.Context, sessionID string) {
	defer close(s.dbWorkerDone)
	var firstErr error
	appendMessage := func(ctx context.Context, msg common.MessageWrapper) {
		if err := s.sessionStore.AppendMessage(ctx, sessionID, msg); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.notifyDBError(notifyCtx, err)
		}
	}

	for {
		select {
		case msg := <-s.msgToDiscChan:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			appendMessage(ctx, msg)
			cancel()
		case <-s.dbWorkerStop:
			for {
				select {
				case msg := <-s.msgToDiscChan:
					drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					appendMessage(drainCtx, msg)
					cancel()
				default:
					s.dbWorkerDone <- firstErr
					return
				}
			}
		}
	}
}

func (s *Session) startDBWorker(ctx context.Context, sessionID string) {
	if s.dbWorkerStarted {
		return
	}
	s.dbWorkerStarted = true
	go s.writeToDbWorker(ctx, sessionID)
}

func (s *Session) handleError(err error, errChan chan error) {
	if errChan != nil {
		select {
		case errChan <- err:
			if s.cancl != nil {
				s.cancl()
			}

		default:
		}
	}

	s.SetNewContext()
}

func (s *Session) stopDBWorker() error {
	if !s.dbWorkerStarted {
		return nil
	}

	close(s.dbWorkerStop)
	err := <-s.dbWorkerDone

	s.dbWorkerStarted = false
	s.dbWorkerStop = make(chan struct{})
	s.dbWorkerDone = make(chan error, 1)
	s.msgToDiscChan = make(chan common.MessageWrapper, 2)

	return err
}

func (s *Session) setErrorChannel(errChan chan error) {
	s.errMu.Lock()
	s.errChan = errChan
	s.errMu.Unlock()
}

func (s *Session) notifyDBError(ctx context.Context, err error) {
	s.errMu.RLock()
	errChan := s.errChan
	s.errMu.RUnlock()
	if errChan == nil {
		return
	}

	select {
	case errChan <- err:
	case <-ctx.Done():
	default:
	}
}

func (s *Session) RetrieveModels(c chan []GroqModelInfo, errChan chan error) {
	go func() {
		res, err := s.groq.getModels(context.Background())
		if err != nil {
			s.handleError(err, errChan)
			return
		}

		var models []GroqModelInfo
		for _, model := range res.Data {
			if model.Active && slices.Contains(model.SupportedFeatures, "tools") && slices.Contains(model.SupportedFeatures, "reasoning") {
				models = append(models, model)
			}
		}

		c <- models
	}()
}

func NewSession() *Session {
	a := agent.NewAgent("Main", 25)
	msgDC := make(chan common.MessageWrapper, 2)

	s := Session{
		groq:          newGroq(),
		agent:         &a,
		msgToDiscChan: msgDC,
		dbWorkerStop:  make(chan struct{}),
		dbWorkerDone:  make(chan error, 1),
	}

	s.SetNewContext()
	store, err := persistence.Open(s.Ctx)
	if err != nil {
		log.Fatal(err)
	}

	s.sessionStore = store
	return &s
}
