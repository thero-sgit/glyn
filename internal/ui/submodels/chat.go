package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/library"
)


type PromptCmd struct {
	Message string
}

type ChatScreen struct {
	viewport  viewport.Model
	history   []string
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen() tea.Model {

	return &ChatScreen{
		viewport: viewport.New(0, 0),
		history: []string{
			library.Burner(0),
		},
	}
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width, c.viewport.Height = msg.Width, msg.Height
		c.history[0] = library.Burner(c.viewport.Width)
		return c, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			}
		}

		c.viewport, cmd = c.viewport.Update(msg)
		return c, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":

		case "up":			

		case "down":
		}

	case PromptCmd:
		c.history = append(c.history, msg.Message)
	}

	return c, nil
}

func (c *ChatScreen) View() string {
	style := lipgloss.NewStyle().
		Width(c.viewport.Width).
		Height(c.viewport.Height)

	if len(c.history) <= 1 {
		style = style.AlignVertical(lipgloss.Center)
	}

	content := style.Render(strings.Join(c.history, "\n"))

	c.viewport.SetContent(content)
	return c.viewport.View()
}
