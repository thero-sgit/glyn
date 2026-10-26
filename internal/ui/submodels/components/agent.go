package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type AgentResponse struct {
	spinner    spinner.Model
	label      string
	HistoryIdx int
	Working    bool

	//
	showReasoning bool

	//
	reasoningBuffer string
	responseBuffer  string
}

func NewAgentResponse(index, width int) *AgentResponse {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return &AgentResponse {
		spinner:    s,
		HistoryIdx: index,
		Working:    true,
	}
}

func (a *AgentResponse) Init() tea.Cmd {
	return nil
}

func (a *AgentResponse) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		var cmd tea.Cmd
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			case zone.Get("thought-process-btn").InBounds(msg):
				a.showReasoning = !a.showReasoning
				return a, cmd
			}
		}

		return a, cmd

	case api.TickMsg50:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(spinner.TickMsg{})

		return a, cmd

	case api.ResponseChunkCmd:
		if msg.Chunk.Reasoning {
			a.reasoningBuffer = a.reasoningBuffer + msg.Chunk.Data
			a.label = a.spinner.View() + " Thinking..."
			return a, nil
		}

		if msg.Chunk.EOS {
			a.Working = false
			a.label = "·"
			r, error := config.Config.MdRenderer.Render(a.responseBuffer)
			if error == nil {
				a.responseBuffer = r
			}

			return a, nil
		}

		a.label = a.spinner.View() + " Working..."
		a.responseBuffer = a.responseBuffer + msg.Chunk.Data
		return a, nil
	}

	return a, nil
}

func (a *AgentResponse) View() string {
	thoughtProcessBtn := library.SubtleStyle.Render("Thought process ") + library.BoldStyle.Render(">")
	thoughtProcessBtn = zone.Mark("thought-process-btn", thoughtProcessBtn)

	reasoning := thoughtProcessBtn
	if a.showReasoning {
		reasoning = lipgloss.JoinVertical(
			lipgloss.Top,
			thoughtProcessBtn,
			library.SubtleStyle.Render(a.reasoningBuffer),
		)
	}

	header := fmt.Sprintf("%s %s", library.BwLabelStyle.Render(" glyn $ "), a.label)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		reasoning,
		a.responseBuffer,
		header,
		"",
	)
}