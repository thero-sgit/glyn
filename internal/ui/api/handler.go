package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai"
)

type Handler struct {
	Session *ai.Session
}

func NewHandler(session *ai.Session) Handler {
	return Handler {
		Session: session,
	}
}

type errMsg struct{ err error }

func errorListener(ctx context.Context, errChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case e, ok := <-errChan:
			if !ok || e == nil {
				return nil
			}
			return errMsg{err: e}
		case <-ctx.Done():
			return errorListener(ctx, errChan)()
		}
	}
}

func AwaitResponse(ctx context.Context, c chan ai.Chunk) tea.Cmd {
	return func() tea.Msg {
		select {
		case completion, ok := <-c:
			if !ok {
				return nil
			}

			return ResponseChunkCmd{Chunk: completion, C: c}
		case <-ctx.Done():
			return nil
		}
	}
}

type sessionInfo struct {
	chunk ai.Chunk
	c     chan ai.Chunk
}

func awaitSessionInfo(ctx context.Context, c chan ai.Chunk) tea.Cmd {
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
	responseChan := make(chan ai.Chunk)
	sessionInfoChan := make(chan ai.Chunk)
	errChan := make(chan error, 2)

	reqCtx := h.Session.Ctx

	var cmds []tea.Cmd

	if len(h.Session.History) < 1 {
		h.Session.NameSession(prompt, sessionInfoChan, errChan)
		cmds = append(cmds, awaitSessionInfo(reqCtx, sessionInfoChan))
	}

	h.Session.NewPrompt(prompt, responseChan, errChan)

	cmds = append(cmds, errorListener(reqCtx, errChan), AwaitResponse(reqCtx, responseChan))

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
