package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type Sessions struct {
	viewport viewport.Model
	chats    []ChatSession
	cursor   int	
}

func (s *Sessions) Init() tea.Cmd {
	return nil
}

func NewSessionsScreen() tea.Model {
	chats := []ChatSession{
		{
			name: "Refactor Auth", 
			changes: [2]string{"19", "3"},
			tokenUsage: "12,032",
			lastUpdated: "2 hours ago",
		},
		{
			name: "Refactor Auth", 
			changes: [2]string{"19", "3"},
			tokenUsage: "12,032",
			lastUpdated: "2 hours ago",
		},
		{
			name: "Refactor Auth", 
			changes: [2]string{"19", "3"},
			tokenUsage: "12,032",
			lastUpdated: "2 hours ago",
		},
	}

	return &Sessions{
		viewport: viewport.New(0, 0),
		chats: chats,
	}
}

func (s *Sessions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.viewport.Width, s.viewport.Height = msg.Width, msg.Height
		return s, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		s.viewport, cmd = s.viewport.Update(msg)
		return s, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":

		case "up":
			if s.cursor > 0 {
				s.cursor--
			} else {
				s.cursor = len(s.chats)-1
			}

			return s, nil

		case "down":
			if s.cursor < len(s.chats)-1 {
				s.cursor++
			} else {
				s.cursor = 0
			}		

			return s, nil
		}
	}

	return s, nil
}

func (s *Sessions) View() string {
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		library.BwLabelStyle.Render(" /sessions "),
		" ",
		config.Config.SanitizedWd,
		" · ",
		library.TokenUsageProgessBar(25.0),
		library.SubtleStyle.MarginLeft(5).Render("press 'esc' to close"),
	)

	var content []string
	for i, chat := range s.chats {
		content = append(content, chat.pretty(s.viewport.Width, i == s.cursor, i == s.cursor))
	}

	s.viewport.SetContent(
		lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().MarginBottom(1).Render(header),
			strings.Join(content, "\n"),
		),
	)

	return s.viewport.View()
}

type ChatSession struct {
	name        string
	changes     [2]string
	tokenUsage  string
	lastUpdated string
}

func (c ChatSession) pretty(width int, highlight, current bool) string {
	header := "\u25CB  " + c.name
	if current {
		header = "\u25CF  " + c.name
	}

	style := lipgloss.NewStyle()
	if highlight {
		style = library.HighlighedStyle
	}

	inf := style.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"   ",
			style.Render("+"+c.changes[0] + " -"+c.changes[1]),
			style.Render(" · " + c.lastUpdated),
			style.Render(" · " + c.tokenUsage + " tokens"),
		),
	)

	return style.Width(width-4).MarginBottom(1).PaddingLeft(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			inf,
		),
	)
}
