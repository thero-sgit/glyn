package submodels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

type Sessions struct {
	Focused bool
}

func (h *Sessions) Init() tea.Cmd {
	return nil
}

func NewSessions() Sessions {
	return Sessions{}
}

func (h *Sessions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// switch msg := msg.(type) {
	// }

	return h, nil
}

func (h *Sessions) View() string {
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		stylesandcomponents.BwLabelStyle.Render(" /sessions "),
		" ",
		config.Config.SanitizedWd,
		stylesandcomponents.SubtleStyle.MarginLeft(5).Render("press 'esc' to close"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
	)
}
