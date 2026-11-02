package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/persistence"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type Home struct {
	viewport viewport.Model
	height   int
	width    int
	chats    []persistence.ChatSession
	cursor   int	
}

func (h *Home) Init() tea.Cmd {
	return nil
}

func NewHomeScreen() tea.Model {
	v := viewport.New(0, 0)
	v.Style.AlignHorizontal(lipgloss.Center)

	return &Home{
		viewport: v,
		chats: persistence.DummySessions(),
	}
}

func (h *Home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width, h.height = msg.Width, msg.Height

		if h.width > 200 {
			h.viewport.Width = h.width/2
		} else {
			h.viewport.Width = h.width - 10
		}

		h.viewport.Height = h.height/2

		return h, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		h.viewport, cmd = h.viewport.Update(msg)
		return h, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":

		case "ctrl+n":
			h.chats = append(h.chats, persistence.NewChatSession("Test Session"))
			return h, nil

		case "up":
			if h.cursor > 0 {
				h.cursor--
			} else {
				h.cursor = len(h.chats)-1
			}

			return h, nil

		case "down":
			if h.cursor < len(h.chats)-1 {
				h.cursor++
			} else {
				h.cursor = 0
			}		

			return h, nil
		}
	}

	return h, nil
}

func (h *Home) View() string {
	var content []string
	for i, chat := range h.chats {
		content = append(content, chat.Pretty(h.viewport.Width, i == h.cursor, i == h.cursor))
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		library.BwLabelStyle.Render(" Recent sessions "),
		" ",
		library.SubtleStyle.Render(config.Config.SanitizedWd),
	)

	header = lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(h.viewport.Width/2).Render(header),
		lipgloss.NewStyle().
			Width(h.viewport.Width/2).
			AlignHorizontal(lipgloss.Right).
			Render(library.SubtleStyle.Render("press 'esc' to close")),
	)

	h.viewport.SetContent(
		lipgloss.JoinVertical(
			lipgloss.Top,
			strings.Join(content, "\n"),
		),
	)

	s := lipgloss.NewStyle().
		Width(h.width).
		AlignHorizontal(lipgloss.Center).
		Render(lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			"",
			h.viewport.View(),
		))

	b := library.Burner(h.width)
	if h.height < 20 {
		b = ""
	}

	return lipgloss.NewStyle().
		Height(h.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				b,
				s,
			),
		)
}