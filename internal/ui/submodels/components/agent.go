package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type AgentResponse struct {
	spinner    spinner.Model
	status     string
	HistoryIdx int
	Working    bool

	showReasoning    bool
	reasoningBuffer  string
	responseBuffer   string
	renderedResponse string
	renderError      string
	toolEvents       map[string]*UIEvent
	toolOrder        []*UIEvent
}

func NewAgentResponse(index, width int) *AgentResponse {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return &AgentResponse{
		spinner:    s,
		status:     "Thinking",
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
			case zone.Get(fmt.Sprintf("%d-thought-process-btn", a.HistoryIdx)).InBounds(msg):
				a.showReasoning = !a.showReasoning
				return a, cmd
			}
		}

		return a, cmd

	case api.TickMsg50:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(spinner.TickMsg{})

		return a, cmd

	case api.TickMsg60:
		var cmds []tea.Cmd
		for _, event := range a.toolOrder {
			if event.Done {
				continue
			}
			_, cmd := event.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case api.ResponseChunkCmd:
		if msg.Chunk.Reasoning {
			a.reasoningBuffer = a.reasoningBuffer + msg.Chunk.Data
			a.status = "Thinking"
			return a, nil
		}

		if msg.Chunk.EOS {
			a.Finish()
			return a, nil
		}

		a.status = "Responding"
		a.responseBuffer = a.responseBuffer + msg.Chunk.Data
		return a, nil

	case agent.Event:
		switch event := msg.(type) {
		case agent.ToolStart:
			if a.toolEvents == nil {
				a.toolEvents = make(map[string]*UIEvent)
			}
			uiEvent := NewUIEvent(a.HistoryIdx, event)
			a.toolEvents[event.ID] = uiEvent
			a.toolOrder = append(a.toolOrder, uiEvent)
			a.status = "Using tools"
		case agent.ToolDone:
			if uiEvent, ok := a.toolEvents[event.ID]; ok {
				uiEvent.Update(event)
				delete(a.toolEvents, event.ID)
				if len(a.toolEvents) == 0 {
					a.status = "Thinking"
				}
			}
		}
		return a, nil
	}

	return a, nil
}

func (a *AgentResponse) Finish() {
	a.Working = false
	a.status = "Complete"
	r, err := config.Config.MdRenderer.Render(a.responseBuffer)
	if err != nil {
		a.renderError = err.Error()
		return
	}
	a.renderedResponse = r
}

func (a *AgentResponse) Fail(err error) {
	a.Working = false
	a.status = "Failed"
	a.renderError = err.Error()
	for _, event := range a.toolOrder {
		if !event.Done {
			event.Done = true
		}
	}
	clear(a.toolEvents)
}

func (a *AgentResponse) View() string {
	status := a.status
	if a.Working {
		status = a.spinner.View() + " " + status
	} else if status == "Complete" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("#02e696")).Render("✓") + " " + status
	}

	header := fmt.Sprintf("%s %s", library.BwLabelStyle.Render(" glyn $ "), library.SubtleStyle.Render(status))
	sections := []string{header}

	if a.reasoningBuffer != "" {
		toggle := "›"
		if a.showReasoning {
			toggle = "⌄"
		}
		thoughtProcessBtn := library.SubtleStyle.Render(toggle + " Thought process")
		thoughtProcessBtn = zone.Mark(fmt.Sprintf("%d-thought-process-btn", a.HistoryIdx), thoughtProcessBtn)
		sections = append(sections, thoughtProcessBtn)

		if a.showReasoning {
			sections = append(sections, library.SubtleStyle.MarginLeft(2).Render(a.reasoningBuffer))
		}
	}

	for _, event := range a.toolOrder {
		sections = append(sections, event.View())
	}

	response := a.responseBuffer
	if a.renderedResponse != "" {
		response = a.renderedResponse
	}
	if response != "" {
		sections = append(sections, response)
	}
	if a.renderError != "" {
		sections = append(sections, library.SubtleStyle.Render("Markdown rendering failed; showing plain text: "+a.renderError))
	}

	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, sections...))
}

func (a *AgentResponse) HasActiveTools() bool {
	return len(a.toolEvents) > 0
}
