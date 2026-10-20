package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)


type PromptCmd struct {
	Message string
}

type ChatScreen struct {
	burner        string
	viewport      viewport.Model
	history       []string
	currentPrompt *components.UserPrompt
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen() tea.Model {

	return &ChatScreen{
		viewport: viewport.New(0, 0),
	}
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width, c.viewport.Height = msg.Width, msg.Height
		c.burner = library.Burner(c.viewport.Width)
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

	case components.TickMsg20:
		c.currentPrompt.Update(msg)
		c.history[0] = c.currentPrompt.View()
		
		return c, components.DoTick20()

	case PromptCmd:
		c.currentPrompt = components.NewUserPrompt(msg.Message)
		c.history = append(c.history, c.currentPrompt.View())

		return c, components.DoTick20()
	}

	return c, nil
}

func (c *ChatScreen) View() string {
	style := lipgloss.NewStyle().
		Width(c.viewport.Width).
		Height(c.viewport.Height)

	if len(c.history) < 1 {
		style = style.AlignVertical(lipgloss.Center)
	}

	content := style.Render(strings.Join(c.history, "\n"))

	c.viewport.SetContent(content)
	return c.viewport.View()
}
