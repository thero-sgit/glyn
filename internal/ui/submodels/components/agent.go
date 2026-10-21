package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ui/library"
)


type TickMsg50 time.Time

func DoTick50() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg50(t)
	})
}

type AgentResponse struct {
	spinner    spinner.Model
	v  string
	HistoryIdx int
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
	switch msg.(type) {
	case TickMsg50:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(spinner.TickMsg{})

		return a, cmd
	}

	return a, nil
}

func (a *AgentResponse) View() string {
	header := fmt.Sprintf("%s %s Working...", library.BwLabelStyle.Render(" glyn $ "), a.spinner.View())

	return header
}