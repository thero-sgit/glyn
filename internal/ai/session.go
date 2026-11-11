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
	"github.com/thero-sgit/glyn/internal/common"
	"github.com/thero-sgit/glyn/internal/persistence"
)

type Session struct {
	Ctx   context.Context
	cancl context.CancelFunc
	groq  *groq
	Wg    sync.WaitGroup
	agent *agent.Agent

	History []agent.Message

	//
	sessionStore    *persistence.Store
	msgToDiscChan   chan agent.MessageWrapper
	dbWorkerStop    chan struct{}
	dbWorkerDone    chan error
	dbWorkerStarted bool
	closeOnce       sync.Once
	closeErr        error
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
		if s.cancl != nil {
			s.cancl()
		}
		s.Wg.Wait()
		if s.dbWorkerStarted {
			close(s.dbWorkerStop)
			s.closeErr = <-s.dbWorkerDone
		}
	})
	return s.closeErr
}

func (s *Session) NameSession(message string, errChan chan error) {
	req := StreamRequest{
		Model: "openai/gpt-oss-20b",
		Messages: []agent.Message{
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
					s.startDBWorker(ctx, c.Id, errChan)
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

func (s *Session) NewPrompt(message string, c chan agent.Chunk, errChan chan error, eventsChan chan<- agent.Event) {
	h := append(s.History, agent.Message{
		Role:    "user",
		Content: message,
	})
	now := time.Now().Format(common.TsFormat)

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

func (s *Session) ContinueSession(sessionId string) ([]agent.Message, error) {
	h, err := s.sessionStore.ListMessagesInSession(s.Ctx, sessionId)
	if err == nil {
		s.History = h
	}

	return h, err
}

func (s *Session) writeToDbWorker(notifyCtx context.Context, sessionID string, errChan chan<- error) {
	defer close(s.dbWorkerDone)
	var firstErr error
	appendMessage := func(ctx context.Context, msg agent.MessageWrapper) {
		if err := s.sessionStore.AppendMessage(ctx, sessionID, msg); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			if errChan != nil {
				select {
				case errChan <- err:
				case <-notifyCtx.Done():
				default:
				}
			}
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

func (s *Session) startDBWorker(ctx context.Context, sessionID string, errChan chan<- error) {
	if s.dbWorkerStarted {
		return
	}
	s.dbWorkerStarted = true
	go s.writeToDbWorker(ctx, sessionID, errChan)
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

func (s *Session) RetrieveModels(c chan []GroqModelInfo, errChan chan error) {
	go func() {
		res, err := s.groq.getModels(s.Ctx)
		if err != nil {
			s.handleError(err, errChan)
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
	msgDC := make(chan agent.MessageWrapper, 2)

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
