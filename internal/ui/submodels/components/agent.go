package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ai/common"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type Run struct {
	ID       string
	Messages []common.Message

	spinner    spinner.Model
	status     string
	HistoryIdx int
	Working    bool
	width      int

	blocks            []*agentResponseBlock
	activeBlock       *agentResponseBlock
	activeMessage     int
	toolEvents        map[string]*UIEvent
	failureError      string
	showSteps         bool
	completedToolCall map[string]bool
	viewDirty         bool
	cachedView        string
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

func NewRun(id string, index, width int, userMessage string) *Run {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return &Run{
		ID:                id,
		Messages:          []common.Message{{Role: "user", Content: userMessage}},
		spinner:           s,
		status:            "Thinking",
		HistoryIdx:        index,
		Working:           true,
		width:             width,
		showSteps:         true,
		activeMessage:     -1,
		completedToolCall: make(map[string]bool),
		viewDirty:         true,
	}
}

func NewHistoricalRun(id string, index, width int, wrappers []common.MessageWrapper) *Run {
	run := NewRun(id, index, width, "")
	run.Messages = make([]common.Message, 0, len(wrappers))
	toolEvents := make(map[string]*UIEvent)
	for _, wrapper := range wrappers {
		message := wrapper.Msg
		messageToolCalls := callMap(message.ToolCalls)
		run.Messages = append(run.Messages, message)
		if message.Role != "assistant" {
			if message.Role == "tool" {
				if uiEvent, ok := toolEvents[message.ToolCallId]; ok {
					start := uiEvent.e.(common.ToolStart)
					uiEvent.Update(common.ToolDone{
						ID:     message.ToolCallId,
						Name:   start.Name,
						Result: message.Content,
					})
					uiEvent.showResult = true
				} else {
					uiEvent := NewUIEvent(index, common.ToolDone{
						ID:     message.ToolCallId,
						Name:   messageToolCalls[message.ToolCallId].Function.Name,
						Result: message.Content,
					})
					uiEvent.Done = true
					uiEvent.showResult = true
					run.blocks = append(run.blocks, &agentResponseBlock{kind: toolBlock, tool: uiEvent})
				}
			}
			continue
		}

		if message.Content != "" {
			block := &agentResponseBlock{kind: responseBlock}
			block.response.WriteString(message.Content)
			run.blocks = append(run.blocks, block)
		}
		for _, call := range message.ToolCalls {
			event := common.ToolStart{ID: call.Id, Name: call.Function.Name, Args: call.Function.Arguments}
			uiEvent := NewUIEvent(index, event)
			if call.Id != "" {
				toolEvents[call.Id] = uiEvent
			}
			run.blocks = append(run.blocks, &agentResponseBlock{kind: toolBlock, tool: uiEvent})
		}
	}
	run.Finish()
	return run
}

func (a *Run) Init() tea.Cmd {
	return nil
}

func (a *Run) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	a.viewDirty = true
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
			a.Messages = append(a.Messages, common.Message{Role: "assistant"})
			a.activeMessage = len(a.Messages) - 1
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
		a.ensureAssistantMessage().Content += msg.Chunk.Data
		return a, nil

	case common.Event:
		switch event := msg.(type) {
		case common.ToolStart:
			if a.toolEvents == nil {
				a.toolEvents = make(map[string]*UIEvent)
			}
			uiEvent := NewUIEvent(a.HistoryIdx, event)
			a.toolEvents[event.ID] = uiEvent
			message := a.ensureAssistantMessage()
			message.ToolCalls = append(message.ToolCalls, common.ToolCall{
				Id:    event.ID,
				Index: len(message.ToolCalls),
				Type:  "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      event.Name,
					Arguments: event.Args,
				},
			})
			a.activeBlock = &agentResponseBlock{kind: toolBlock, tool: uiEvent}
			a.blocks = append(a.blocks, a.activeBlock)
			a.status = "Using tools"
		case common.ToolDone:
			if uiEvent, ok := a.toolEvents[event.ID]; ok {
				uiEvent.Update(event)
				delete(a.toolEvents, event.ID)
				if !a.completedToolCall[event.ID] {
					a.Messages = append(a.Messages, common.Message{
						Role:       "tool",
						Content:    event.Result,
						ToolCallId: event.ID,
					})
					a.completedToolCall[event.ID] = true
				}
				if len(a.toolEvents) == 0 {
					a.status = "Thinking"
				}
			}
		}
		return a, nil
	}

	return a, nil
}

func (a *Run) Finish() {
	a.viewDirty = true
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

func (a *Run) Fail(err error) {
	a.viewDirty = true
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

func (a *Run) View() string {
	if !a.viewDirty {
		return a.cachedView
	}

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
					sections = append(sections, lipgloss.NewStyle().
						Width(max(1, a.width-4)).
						MarginLeft(4).
						Render(block.reasoning.String()))
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

	a.cachedView = lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, sections...))
	a.viewDirty = false
	return a.cachedView
}

func (a *Run) HasActiveTools() bool {
	return len(a.toolEvents) > 0
}

func (a *Run) SetWidth(width int) {
	if a.width == width {
		return
	}
	a.width = width
	a.viewDirty = true
}

func (a *Run) ensureBlock(kind agentResponseBlockKind) *agentResponseBlock {
	if a.activeBlock != nil && a.activeBlock.kind == kind {
		return a.activeBlock
	}
	a.activeBlock = &agentResponseBlock{kind: kind}
	a.blocks = append(a.blocks, a.activeBlock)
	return a.activeBlock
}

func (a *Run) ensureAssistantMessage() *common.Message {
	if a.activeMessage < 0 || a.activeMessage >= len(a.Messages) || a.Messages[a.activeMessage].Role != "assistant" {
		a.Messages = append(a.Messages, common.Message{Role: "assistant"})
		a.activeMessage = len(a.Messages) - 1
	}
	return &a.Messages[a.activeMessage]
}

func (a *Run) stepsZone() string {
	return fmt.Sprintf("%d-steps", a.HistoryIdx)
}

func (a *Run) reasoningZone(index int) string {
	return fmt.Sprintf("%d-thought-process-%d", a.HistoryIdx, index)
}

func callMap(calls []common.ToolCall) map[string]common.ToolCall {
	cMap := make(map[string]common.ToolCall)
	for i := range calls {
		cMap[calls[i].Id] = calls[i]
	}
	return cMap
}
