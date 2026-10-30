package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai"
	"github.com/thero-sgit/glyn/internal/ai/agent"
)

type Handler struct {
	Session *ai.Session
}

func NewHandler(session *ai.Session) Handler {
	return Handler {
		Session: session,
	}
}

func errorListener(ctx context.Context, ch <-chan error) tea.Cmd {
    return func() tea.Msg {
        select {
        case err, ok := <-ch:
            if !ok {
                return nil
            }
            return ErrorMsg{err, ch}
        case <-ctx.Done():
            return nil
        }
    }
}

func AwaitEvent(ctx context.Context, ch <-chan agent.Event) tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-ch:
			if !ok {
				return nil
			}

			return EventMsg{Event: event, Ctx: ctx, C: ch}
		case <-ctx.Done():
			return nil
		}
	}
}

func AwaitResponse(ctx context.Context, ch chan agent.Chunk) tea.Cmd {
	return func() tea.Msg {
		select {
		case completion, ok := <-ch:
			if !ok {
				return nil
			}

			return ResponseChunkCmd{Chunk: completion, C: ch}
		case <-ctx.Done():
			return nil
		}
	}
}

type sessionInfo struct {
	chunk agent.Chunk
	c     chan agent.Chunk
}

func awaitSessionInfo(ctx context.Context, c chan agent.Chunk) tea.Cmd {
	return func() tea.Msg {
		select {
		case chunk, ok := <-c:
			if !ok {
				return nil
			}
			return sessionInfo{chunk: chunk, c: c}
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *Handler) HandlePrompt(prompt string) tea.Cmd {
	responseChan    := make(chan agent.Chunk)
	sessionInfoChan := make(chan agent.Chunk)
	eventsChan      := make(chan agent.Event)
	errChan         := make(chan error, 2)

	reqCtx := h.Session.Ctx

	var cmds []tea.Cmd

	if len(h.Session.History) < 1 {
		h.Session.NameSession(prompt, sessionInfoChan, errChan)
		cmds = append(cmds, awaitSessionInfo(reqCtx, sessionInfoChan))
	}

	h.Session.NewPrompt(prompt, responseChan, errChan, eventsChan)

	cmds = append(cmds, errorListener(reqCtx, errChan), AwaitEvent(reqCtx, eventsChan), AwaitResponse(reqCtx, responseChan))

	return tea.Batch(cmds...)
}

func awaitModels(ctx context.Context, c chan []ai.GroqModelInfo) tea.Cmd {
	return func() tea.Msg {
		select {
		case m, ok := <-c:
			if !ok {
				return nil
			}

			return ModelsCmd{M: m, C: c}
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *Handler) HandleGetModels() tea.Cmd {
	responseChan := make(chan []ai.GroqModelInfo)
	errChan := make(chan error, 2)

	h.Session.RetrieveModels(responseChan, errChan)
	return tea.Batch(errorListener(h.Session.Ctx, errChan), awaitModels(h.Session.Ctx, responseChan))
}
