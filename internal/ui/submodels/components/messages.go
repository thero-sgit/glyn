package components

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type TickMsg20 time.Time

func DoTick20() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg20(t)
	})
}

type UserPrompt struct {
	message   string
	waveIndex int
}

func (u *UserPrompt) Init() tea.Cmd {
	return nil
}

func NewUserPrompt(message string) *UserPrompt {
	return &UserPrompt{
		message: message,
	}
}

func (u *UserPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg20:
		u.waveIndex++
	}

	return u, nil
}

func (u *UserPrompt) View() string {
	var newPretty strings.Builder
	for i, v := range u.message {
		phase := float64(u.waveIndex)*0.2 - float64(i)*0.4
		offset := math.Sin(phase)

		style := library.ColorForOffset(offset)
		newPretty.WriteString(style.Render(string(v)))
	}

	return  lipgloss.NewStyle().
		MarginBottom(1).
		PaddingRight(1).
		Render(
			library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().PaddingLeft(1).Render(newPretty.String()),
		)
}

func (u *UserPrompt) Pretty() string {
	return u.View()
}