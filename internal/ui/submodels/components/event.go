package components

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type UIEvent struct {
	HistoryIdx int
	Done       bool
	showResult bool
	//
	spinner    spinner.Model
	e          agent.Event
	args       string
	viewDirty  bool
	cachedView string
}

func NewUIEvent(index int, event agent.Event) *UIEvent {
	s := spinner.New()
	s.Spinner = spinner.Pulse

	return &UIEvent{
		HistoryIdx: index,
		spinner:    s,
		e:          event,
		args:       toolStartArgs(event),
		viewDirty:  true,
	}
}

func (u *UIEvent) Init() tea.Cmd {
	return nil
}

func (u *UIEvent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	u.viewDirty = true
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if result, ok := u.e.(agent.ToolDone); ok && result.Result != "" &&
				zone.Get(u.resultZone()).InBounds(msg) {
				u.showResult = !u.showResult
			}
		}

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
	if !u.viewDirty {
		return u.cachedView
	}

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

	line := fmt.Sprintf("  %s %s", status, name)
	if detail != "" {
		line += " " + library.SubtleStyle.Render(detail)
	}
	if result, ok := u.e.(agent.ToolDone); ok && result.Result != "" {
		toggle := "› Result"
		if u.showResult {
			toggle = "⌄ Result"
		}
		resultToggle := zone.Mark(u.resultZone(), library.SubtleStyle.MarginLeft(4).Render(toggle))
		line = lipgloss.JoinVertical(lipgloss.Left, line, resultToggle)
		if u.showResult {
			resultLine := toolResultView(eventName(u.e), result.Result)
			line = lipgloss.JoinVertical(lipgloss.Left, line, resultLine)
		}
	}
	u.cachedView = lipgloss.NewStyle().MarginBottom(1).Render(line)
	u.viewDirty = false
	return u.cachedView
}

func toolStartArgs(event agent.Event) string {
	if start, ok := event.(agent.ToolStart); ok {
		return start.Args
	}
	return ""
}

func (u *UIEvent) resultZone() string {
	var id string
	switch event := u.e.(type) {
	case agent.ToolStart:
		id = event.ID
	case agent.ToolDone:
		id = event.ID
	}
	return fmt.Sprintf("%d-tool-result-%s", u.HistoryIdx, id)
}

func toolLabel(name string) string {
	switch name {
	case "read_file":
		return "Read"
	case "web_search":
		return "Search"
	default:
		return strings.ReplaceAll(name, "_", " ")
	}
}

func toolDetail(value string) string {
	var args struct {
		Path  string `json:"path"`
		Query string `json:"query"`
	}
	if json.Unmarshal([]byte(value), &args) == nil {
		if args.Path != "" {
			return "(" + args.Path + ")"
		}
		if args.Query != "" {
			return "(" + args.Query + ")"
		}
	}
	return truncateDetail(value, 120)
}

func eventName(event agent.Event) string {
	switch event := event.(type) {
	case agent.ToolStart:
		return event.Name
	case agent.ToolDone:
		return event.Name
	default:
		return ""
	}
}

func toolResultView(toolName, result string) string {
	if toolName != "web_search" {
		return library.SubtleStyle.MarginLeft(6).Render("⎿ " + truncateDetail(result, 320))
	}

	const resultPrefix = "<search_results>"
	const resultSuffix = "<search_results>"
	payload := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(result), resultPrefix), resultSuffix))
	var response struct {
		WebPages struct {
			Values []struct {
				Name        string `json:"name"`
				URL         string `json:"url"`
				DisplayURL  string `json:"displayUrl"`
				Snippet     string `json:"snippet"`
				Text        string `json:"text"`
				DatePublish string `json:"datePublished"`
			} `json:"value"`
		} `json:"webPages"`
	}
	if err := json.Unmarshal([]byte(payload), &response); err != nil || len(response.WebPages.Values) == 0 {
		return library.SubtleStyle.MarginLeft(6).Render("⎿ " + truncateDetail(result, 320))
	}

	var entries []string
	for index, item := range response.WebPages.Values {
		if index == 5 {
			entries = append(entries, library.SubtleStyle.Render(fmt.Sprintf("… and %d more results", len(response.WebPages.Values)-index)))
			break
		}
		title := item.Name
		if title == "" {
			title = item.URL
		}
		entry := lipgloss.JoinVertical(
			lipgloss.Left,
			library.HighlighedStyle.Render(fmt.Sprintf("%d. %s", index+1, title)),
			library.SubtleStyle.MarginLeft(3).Render(firstNonEmpty(item.DisplayURL, item.URL)),
		)
		snippet := firstNonEmpty(item.Snippet, item.Text)
		if snippet != "" {
			entry = lipgloss.JoinVertical(lipgloss.Left, entry, library.SubtleStyle.MarginLeft(3).Render(truncateDetail(snippet, 180)))
		}
		if item.DatePublish != "" {
			entry = lipgloss.JoinVertical(lipgloss.Left, entry, library.SubtleStyle.MarginLeft(3).Render(item.DatePublish))
		}
		entries = append(entries, entry)
	}

	return lipgloss.NewStyle().MarginLeft(6).Render(lipgloss.JoinVertical(lipgloss.Left, entries...))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func truncateDetail(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit]) + "…"
	}
	return value
}
