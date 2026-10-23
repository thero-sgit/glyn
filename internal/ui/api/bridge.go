package api

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai"
)

type Handler struct {
	session *ai.Session
}

func NewHandler(session *ai.Session) Handler {
	return Handler {
		session: session,
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

type response struct {
	chunk ai.Chunk
	c     chan ai.Chunk
}

func awaitResponse(ctx context.Context, c chan ai.Chunk) tea.Cmd {
	return func() tea.Msg {
		select {
		case completion, ok := <-c:
			if !ok {
				return nil
			}

			return response{chunk: completion, c: c}
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

func (h *Handler) handlePrompt(prompt string) tea.Cmd {
	responseChan := make(chan ai.Chunk)
	sessionInfoChan := make(chan ai.Chunk)
	errChan := make(chan error, 2)

	reqCtx := h.session.Ctx

	var cmds []tea.Cmd

	if len(h.session.History) < 1 {
		h.session.NameSession(prompt, sessionInfoChan, errChan)
		cmds = append(cmds, awaitSessionInfo(reqCtx, sessionInfoChan))
	}

	h.session.NewPrompt(prompt, responseChan, errChan)

	cmds = append(cmds, errorListener(reqCtx, errChan), awaitResponse(reqCtx, responseChan))

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

func (h *Handler) handleGetModels() tea.Cmd {
	responseChan := make(chan []ai.GroqModelInfo)
	errChan := make(chan error, 2)

	h.session.RetrieveModels(responseChan, errChan)

	return tea.Batch(errorListener(h.session.Ctx, errChan), awaitModels(h.session.Ctx, responseChan))
}
