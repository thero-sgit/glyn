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

type SentCmd  struct {}
type ErrorCmd struct {}

type UserPrompt struct {
	message    string
	waveIndex  int
	pretty     string
	HistoryIdx int
}

func (u *UserPrompt) Init() tea.Cmd {
	return nil
}

func NewUserPrompt(message string, index int) *UserPrompt {
	return &UserPrompt{
		message:    message,
		HistoryIdx: index,
	}
}

func (u *UserPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg20:
		u.waveIndex++
		u.animate()

	case SentCmd:
		u.sent()

	case ErrorCmd:
		u.error()
	}

	return u, nil
}

func (u *UserPrompt) View() string {
	return u.pretty
}

func (u *UserPrompt) animate() {
	var newPretty strings.Builder

	for i, v := range u.message {
		phase := float64(u.waveIndex) * 0.2 - float64(i) * 0.4
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
		library.RetryButton(),
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