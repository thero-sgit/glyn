package submodels

import (
	tea "github.com/charmbracelet/bubbletea"
)


type ChatScreen struct {
	width    int
	chats    []ChatSession
	cursor   int
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen() tea.Model {
	return &ChatScreen{}
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":

		case "up":			

		case "down":		
		}
	}

	return c, nil
}

func (c *ChatScreen) View() string {
	return ""
}
