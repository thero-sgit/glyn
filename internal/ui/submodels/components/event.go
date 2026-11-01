package components

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type UIEvent struct {
	HistoryIdx int
	Done       bool
	//
	spinner spinner.Model
	e       agent.Event
	args    string
}

func NewUIEvent(index int, event agent.Event) *UIEvent {
	s := spinner.New()
	s.Spinner = spinner.Pulse

	return &UIEvent{
		HistoryIdx: index,
		spinner:    s,
		e:          event,
		args:       toolStartArgs(event),
	}
}

func (u *UIEvent) Init() tea.Cmd {
	return nil
}

func (u *UIEvent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case api.TickMsg60:
		var cmd tea.Cmd
		u.spinner, cmd = u.spinner.Update(spinner.TickMsg{})

		return u, cmd

	case agent.Event:
		switch e := msg.(type) {
		case agent.ToolDone:
			if start, ok := u.e.(agent.ToolStart); ok && start.ID == e.ID {
				u.e = e
				u.Done = true
			}
		}

		return u, nil
	}

	return u, nil
}

func (u *UIEvent) View() string {
	var name, detail string
	switch event := u.e.(type) {
	case agent.ToolStart:
		name = toolLabel(event.Name)
		detail = toolDetail(u.args)
	case agent.ToolDone:
		name = toolLabel(event.Name)
		detail = toolDetail(u.args)
	default:
		name = u.e.Pretty()
	}

	status := u.spinner.View()
	if u.Done {
		status = "✓"
	}

	line := fmt.Sprintf("%s %s", status, name)
	if detail != "" {
		line += " " + library.SubtleStyle.Render(detail)
	}
	if result, ok := u.e.(agent.ToolDone); ok && result.Result != "" {
		resultLine := library.SubtleStyle.MarginLeft(2).Render("⎿ " + truncateDetail(result.Result, 320))
		line = lipgloss.JoinVertical(lipgloss.Left, line, resultLine)
	}
	return lipgloss.NewStyle().MarginBottom(1).Render(line)
}

func toolStartArgs(event agent.Event) string {
	if start, ok := event.(agent.ToolStart); ok {
		return start.Args
	}
	return ""
}

func toolLabel(name string) string {
	switch name {
	case "read_file":
		return "Read"
	default:
		return strings.ReplaceAll(name, "_", " ")
	}
}

func toolDetail(value string) string {
	var args struct {
		Path string `json:"path"`
	}
	if json.Unmarshal([]byte(value), &args) == nil && args.Path != "" {
		return "(" + args.Path + ")"
	}
	return truncateDetail(value, 120)
}

func truncateDetail(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit]) + "…"
	}
	return value
}
