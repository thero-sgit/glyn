package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type AgentResponse struct {
	spinner    spinner.Model
	HistoryIdx int

	//
	responseBuffer string
}

func NewAgentResponse(index int) *AgentResponse {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return &AgentResponse {
		spinner:    s,
		HistoryIdx: index,
	}
}

func (a *AgentResponse) Init() tea.Cmd {
	return nil
}

func (a *AgentResponse) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case api.TickMsg50:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(spinner.TickMsg{})

		return a, cmd

	case api.ResponseChunkCmd:
		a.responseBuffer = a.responseBuffer + msg.Chunk.Data
		return a, nil
	}

	return a, nil
}

func (a *AgentResponse) View() string {
	header := fmt.Sprintf("%s %s Working...", library.BwLabelStyle.Render(" glyn $ "), a.spinner.View())

	return lipgloss.JoinVertical(
		lipgloss.Top,
		a.responseBuffer,
		"",
		header,
	)
}