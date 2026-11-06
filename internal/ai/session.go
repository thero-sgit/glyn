package ai

import (
	"context"
	"errors"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/persistence"
)

type Session struct {
    Ctx    context.Context
    cancl  context.CancelFunc
    groq   *groq
    Wg     sync.WaitGroup
    agent  *agent.Agent

    History       []agent.Message
    chatSession   *persistence.ChatSession

}

func (s *Session) SetNewContext() {
    if s.cancl != nil {
        s.cancl()
    }

    ctx, cancel := context.WithCancel(context.Background())
    s.Ctx = ctx
    s.cancl = cancel
}

func (s *Session) Close() {
    if s.cancl != nil {
        s.cancl()
    }
    s.Wg.Wait()
}

func (s *Session) NameSession(message string, errChan chan error) {
    req := StreamRequest {
        Model: "openai/gpt-oss-20b",
        Messages: []agent.Message{
            {Role: "system", Content: "Generate a 3-5 word title for this prompt. Return ONLY the title. MUST BE 3 TO 5 WORDS!"},
            {Role: "user", Content: message},
        },
        Stream: true,
    }

    s.Wg.Add(1)
    
    var name string
    go func() {
        defer s.Wg.Done()

        requestCtx, cancel := context.WithCancel(s.Ctx)
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
                    s.chatSession = persistence.NewChatSession(name).New()
                    return
                }
                errChan <- err
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

    s.Wg.Add(1)
    go func() {
        defer s.Wg.Done()
        defer close(c)
        defer close(eventsChan)

        requestCtx, cancel := context.WithCancel(s.Ctx)
        defer cancel()

        h, err := s.agent.Run(requestCtx, s.groq.handlePrompt, h, c, eventsChan)
        if err != nil {
            s.handleError(err, errChan)
            return
        }
        s.History = h
    }()
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

func (s *Session) RetrieveModels(c chan []GroqModelInfo, errChan chan error){
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

    s := Session{
        groq:  newGroq(),
        agent: &a,
    }
    s.SetNewContext()

    return &s
}