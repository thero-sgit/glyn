package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/persistence"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type Sessions struct {
	viewport viewport.Model
	chats    []persistence.ChatSession
	cursor   int

	//
	projectKey string
}

func (s *Sessions) Init() tea.Cmd {
	return nil
}

func NewSessionsScreen() tea.Model {
	s := &Sessions{
		viewport: viewport.New(0, 0),
		projectKey: config.Config.ProjectKey,
	}

	// s.loadChatSessions()
	return s
}

func (s *Sessions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.viewport.Width, s.viewport.Height = msg.Width, msg.Height
		return s, nil

	case api.SessionsReadMsg:
		s.chats = msg.Sessions
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
		content = append(content, chat.Pretty(s.viewport.Width, i == s.cursor, i == s.cursor))
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

// func (s *Sessions) loadChatSessions() {
// 	destination, err := os.ReadDir(filepath.Join(config.Config.DataBaseDir, s.projectKey))
// 	if err != nil {
// 		return
// 	}

// 	var entries []persistence.ChatSession
// 	for _, e := range destination {
// 		id := strings.Split(e.Name(), ".")[0]
// 		if e, err := persistence.NewChatSessionFromId(s.projectKey, id); err == nil {
// 			entries = append(entries, *e)
// 		}
// 	}

// 	s.chats = entries
// }