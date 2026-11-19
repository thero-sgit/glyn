package api

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai"
	"github.com/thero-sgit/glyn/internal/ai/common"
	"github.com/thero-sgit/glyn/internal/persistence"
)

type SentCmd struct{}

type ErrorMsg struct {
	E error
	C <-chan error
}

type ModelsCmd struct {
	M []ai.GroqModelInfo
	C chan []ai.GroqModelInfo
}

type ModelsLoadingTick struct{}

func DoModelsLoadingTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return ModelsLoadingTick{}
	})
}

type SessionLoadingTick struct{}

func DoSessionLoadingTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return SessionLoadingTick{}
	})
}

type TickMsg20 time.Time

func DoTick20() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg20(t)
	})
}

type TickMsg50 time.Time

func DoTick50() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg50(t)
	})
}

type TickMsg60 time.Time

func DoTick60() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg60(t)
	})
}

type PromptCmd struct {
	Message string
}

type ResponseChunkCmd struct {
	Chunk common.Chunk
	C     chan common.Chunk
}

type EventCmd struct {
	Event common.Event
	Ctx   context.Context
	C     <-chan common.Event
}

type SessionsReadMsg struct {
	Sessions []persistence.ChatSession
	Error    error
}

type GetSessionHistory struct {
	SessionId string
}

func GoGetSessionHistory(id string) tea.Cmd {
	return func() tea.Msg {
		return GetSessionHistory{SessionId: id}
	}
}

type ContinueSessionMsg struct {
	Messages []common.MessageWrapper
	Error    error
}
