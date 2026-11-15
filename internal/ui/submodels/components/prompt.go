package components

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type UserPrompt struct {
	message    string
	waveIndex  int
	pretty     string
	HistoryIdx int
	IsSent     bool
	Errored    bool
}

func NewUserPrompt(message string, index int) *UserPrompt {
	prompt := &UserPrompt{
		message:    message,
		HistoryIdx: index,
	}
	prompt.animate()
	return prompt
}

func (u *UserPrompt) Init() tea.Cmd {
	return nil
}

func (u *UserPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case api.TickMsg20:
		u.waveIndex++
		u.animate()

	case api.SentCmd:
		u.IsSent = true
		u.sent()

	case api.ErrorMsg:
		u.Errored = true
		u.error()
	}

	return u, nil
}

func (u *UserPrompt) View() string {
	return u.pretty
}

func (u *UserPrompt) RetryClicked(msg tea.MouseMsg) (string, bool) {
	if !u.Errored || msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return "", false
	}
	if !zone.Get(u.retryZone()).InBounds(msg) {
		return "", false
	}
	return u.message, true
}

func (u *UserPrompt) animate() {
	var newPretty strings.Builder

	for i, v := range u.message {
		phase := float64(u.waveIndex)*0.2 - float64(i)*0.4
		offset := math.Sin(phase)

		style := library.ColorForOffset(offset)
		newPretty.WriteString(style.Render(string(v)))
	}

	u.pretty = lipgloss.NewStyle().
		MarginBottom(1).
		PaddingRight(1).
		Render(
			library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().PaddingLeft(1).Render(newPretty.String()),
		)
}

func (u *UserPrompt) sent() {
	p := library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().PaddingLeft(1).Render(u.message)

	u.pretty = lipgloss.NewStyle().
		MarginBottom(1).
		PaddingRight(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				p,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#02e696")).PaddingRight(2).Render("\u2713"),
			),
		)
}

func (u *UserPrompt) error() {
	p := library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().Strikethrough(true).PaddingLeft(1).Render(u.message)

	ed := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().MarginRight(4).Render(
			library.SubtleStyle.Render("failed to send"),
		),
		library.RetryButton(u.retryZone()),
	)

	u.pretty = lipgloss.NewStyle().
		MarginBottom(1).
		PaddingRight(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				p,
				ed,
			),
		)
}

func (u *UserPrompt) retryZone() string {
	return fmt.Sprintf("%d-retry-button", u.HistoryIdx)
}
