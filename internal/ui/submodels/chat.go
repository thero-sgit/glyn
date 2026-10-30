package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

type ChatScreen struct {
	viewport      viewport.Model
	history       []string

	//
	handler *api.Handler

	//
	currentPrompt   *components.UserPrompt
	currentResponse *components.AgentResponse
	currentEvent    *components.UIEvent
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen(h *api.Handler) tea.Model {
	return &ChatScreen{
		viewport: viewport.New(0, 0),
		handler: h,
	}
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width, c.viewport.Height = msg.Width, msg.Height

		if len(c.history) == 0 {
			c.history = append(c.history, library.Burner(c.viewport.Width))
		} else {
			c.history[0] = library.Burner(c.viewport.Width)
		}

		config.Config.SetRenderer(c.viewport.Width)

		return c, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			default:
				if cr := c.currentResponse; cr != nil {
					cr := *cr
					updated, uCmd := cr.Update(msg)
					c.currentResponse = updated.(*components.AgentResponse)
					c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
					cmd = uCmd
				}
				
				return c, cmd
			}
		}

		c.viewport, cmd = c.viewport.Update(msg)
		return c, cmd

	case tea.KeyMsg:
		switch msg.String() {

		case "up":

		case "down":
		}

	case api.ErrorMsg:
		if c.currentPrompt != nil {
			c.currentPrompt.Update(msg)
		}		
		c.history = append(c.history, msg.E.Error())

		return c, nil

	case api.PromptCmd:
		c.addNewPropmt(msg.Message)
		cmd := c.handler.HandlePrompt(msg.Message)

		return c, tea.Batch(cmd, api.DoTick20())

	case api.TickMsg20:
		c.history[c.currentPrompt.HistoryIdx] = c.currentPrompt.View()

		if c.currentPrompt.Errored {
			return c, nil
		}

		if c.currentPrompt.IsSent {
			return c, api.DoTick50()
		}

		c.currentPrompt.Update(msg)
		return c, api.DoTick20()

	case api.TickMsg50:
		c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()

		if c.currentResponse.Working {
			c.currentResponse.Update(msg)
			return c, api.DoTick50()
		}	

		return c, nil

	case api.TickMsg60:
		c.currentEvent.Update(msg)
		c.history[c.currentEvent.HistoryIdx] = c.currentEvent.View()

		if c.currentEvent.Done {
			return c, nil
		}

		return c, api.DoTick60()

	case api.ResponseChunkCmd:
		if !c.currentPrompt.IsSent {
			c.currentPrompt.Update(api.SentCmd{})
			c.addNewResponse()
		}
		
		c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()

		if msg.Chunk.EOS {
			c.currentResponse.Update(msg)
			return c, nil
		}

		c.viewport.GotoBottom()
		c.currentResponse.Update(msg)
		return c, api.AwaitResponse(c.handler.Session.Ctx, msg.C)

	case api.EventMsg:
		switch msg.Event.(type) {
		case agent.ToolStart:
			c.addNewEvent(msg.Event)
			c.currentEvent.Update(msg.Event)
			c.history[c.currentEvent.HistoryIdx] = c.currentEvent.View()

			return c, tea.Batch(api.DoTick60(), api.AwaitEvent(msg.Ctx, msg.C))

		case agent.ToolDone:
			c.currentEvent.Done = true
			c.currentEvent.Update(msg.Event)
			c.addNewResponse()

			return c, api.AwaitEvent(msg.Ctx, msg.C)
		}		
	}

	return c, nil
}

func (c *ChatScreen) View() string {
	style := lipgloss.NewStyle().
		Width(c.viewport.Width).
		Height(c.viewport.Height)

	if len(c.history) == 1 {
		style = style.AlignVertical(lipgloss.Center)
	}

	content := style.Render(strings.Join(c.history, "\n"))

	c.viewport.SetContent(content)
	return c.viewport.View()
}

func (c *ChatScreen) addNewPropmt(message string) {
	c.currentPrompt = components.NewUserPrompt(message, len(c.history))
	c.history = append(c.history, c.currentPrompt.View())
}

func (c *ChatScreen) addNewResponse() {
	c.currentResponse = components.NewAgentResponse(len(c.history), c.viewport.Width)
	c.history = append(c.history, c.currentResponse.View())
}

func (c *ChatScreen) addNewEvent(event agent.Event) {
	c.currentEvent = components.NewUIEvent(len(c.history), event)
	c.history = append(c.history, c.currentEvent.View())
}