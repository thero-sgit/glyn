package components

import (
	"fmt"
	"strings"

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

	blocks       []*agentResponseBlock
	activeBlock  *agentResponseBlock
	toolEvents   map[string]*UIEvent
	failureError string
	showSteps    bool
}

type agentResponseBlockKind uint8

const (
	reasoningBlock agentResponseBlockKind = iota
	responseBlock
	toolBlock
)

type agentResponseBlock struct {
	kind             agentResponseBlockKind
	reasoning        strings.Builder
	response         strings.Builder
	renderedResponse string
	renderError      string
	showReasoning    bool
	tool             *UIEvent
}

func NewAgentResponse(index, width int) *AgentResponse {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return &AgentResponse{
		spinner:    s,
		status:     "Thinking",
		HistoryIdx: index,
		Working:    true,
		showSteps:  true,
	}
}

func (a *AgentResponse) Init() tea.Cmd {
	return nil
}

func (a *AgentResponse) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if zone.Get(a.stepsZone()).InBounds(msg) {
				a.showSteps = !a.showSteps
				return a, nil
			}
			for index, block := range a.blocks {
				if block.kind == reasoningBlock && a.showSteps &&
					zone.Get(a.reasoningZone(index)).InBounds(msg) {
					block.showReasoning = !block.showReasoning
					return a, nil
				}
				if block.kind != toolBlock {
					continue
				}
				updated, cmd := block.tool.Update(msg)
				block.tool = updated.(*UIEvent)
				if zone.Get(block.tool.resultZone()).InBounds(msg) {
					return a, cmd
				}
			}
		}
		return a, nil

	case api.TickMsg50:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(spinner.TickMsg{})

		return a, cmd

	case api.TickMsg60:
		var cmds []tea.Cmd
		for _, block := range a.blocks {
			if block.kind != toolBlock || block.tool.Done {
				continue
			}
			_, cmd := block.tool.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case api.ResponseChunkCmd:
		if msg.Chunk.TurnStart {
			a.activeBlock = nil
			a.status = "Thinking"
			return a, nil
		}
		if msg.Chunk.Reasoning {
			block := a.ensureBlock(reasoningBlock)
			block.reasoning.WriteString(msg.Chunk.Data)
			a.status = "Thinking"
			return a, nil
		}

		if msg.Chunk.EOS {
			a.Finish()
			return a, nil
		}

		if msg.Chunk.Data != "" && len(a.toolEvents) == 0 {
			a.showSteps = false
		}
		a.status = "Responding"
		block := a.ensureBlock(responseBlock)
		block.response.WriteString(msg.Chunk.Data)
		return a, nil

	case agent.Event:
		switch event := msg.(type) {
		case agent.ToolStart:
			if a.toolEvents == nil {
				a.toolEvents = make(map[string]*UIEvent)
			}
			uiEvent := NewUIEvent(a.HistoryIdx, event)
			a.toolEvents[event.ID] = uiEvent
			a.activeBlock = &agentResponseBlock{kind: toolBlock, tool: uiEvent}
			a.blocks = append(a.blocks, a.activeBlock)
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
	for _, block := range a.blocks {
		if block.kind != responseBlock || block.renderedResponse != "" || block.response.Len() == 0 {
			continue
		}
		r, err := config.Config.MdRenderer.Render(block.response.String())
		if err != nil {
			block.renderError = err.Error()
		} else {
			block.renderedResponse = r
		}
	}
}

func (a *AgentResponse) Fail(err error) {
	a.Working = false
	a.status = "Failed"
	a.failureError = err.Error()
	for _, block := range a.blocks {
		if block.kind == toolBlock && !block.tool.Done {
			block.tool.Done = true
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

	stepCount := 0
	for _, block := range a.blocks {
		if block.kind == reasoningBlock || block.kind == toolBlock {
			stepCount++
		}
	}
	if stepCount > 0 {
		toggle := "›"
		if a.showSteps {
			toggle = "⌄"
		}
		stepsLabel := fmt.Sprintf("%s Steps (%d)", toggle, stepCount)
		sections = append(sections, zone.Mark(a.stepsZone(), library.SubtleStyle.Render(stepsLabel)))
	}

	for index, block := range a.blocks {
		switch block.kind {
		case reasoningBlock:
			if a.showSteps {
				toggle := "›"
				if block.showReasoning {
					toggle = "⌄"
				}
				label := library.SubtleStyle.Render("  " + toggle + " Thought process")
				sections = append(sections, zone.Mark(a.reasoningZone(index), label))
				if block.showReasoning {
					sections = append(sections, library.SubtleStyle.MarginLeft(4).Render(block.reasoning.String()))
				}
			}
		case toolBlock:
			if a.showSteps {
				sections = append(sections, block.tool.View())
			}
		case responseBlock:
			response := block.response.String()
			if block.renderedResponse != "" {
				response = block.renderedResponse
			}
			if response != "" {
				sections = append(sections, response)
			}
			if block.renderError != "" {
				sections = append(sections, library.SubtleStyle.Render("Markdown rendering failed; showing plain text: "+block.renderError))
			}
		}
	}
	if a.failureError != "" {
		sections = append(sections, library.SubtleStyle.Render(a.failureError))
	}

	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, sections...))
}

func (a *AgentResponse) HasActiveTools() bool {
	return len(a.toolEvents) > 0
}

func (a *AgentResponse) ensureBlock(kind agentResponseBlockKind) *agentResponseBlock {
	if a.activeBlock != nil && a.activeBlock.kind == kind {
		return a.activeBlock
	}
	a.activeBlock = &agentResponseBlock{kind: kind}
	a.blocks = append(a.blocks, a.activeBlock)
	return a.activeBlock
}

func (a *AgentResponse) stepsZone() string {
	return fmt.Sprintf("%d-steps", a.HistoryIdx)
}

func (a *AgentResponse) reasoningZone(index int) string {
	return fmt.Sprintf("%d-thought-process-%d", a.HistoryIdx, index)
}
