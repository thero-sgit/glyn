package api

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai"
)

type SentCmd  struct {}
type ErrorCmd struct {}

type ModelsCmd struct {
	M []ai.GroqModelInfo
	C chan []ai.GroqModelInfo
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

type PromptCmd struct {
	Message string
}


type ResponseChunkCmd struct {
	Chunk ai.Chunk
	C     chan ai.Chunk
}
