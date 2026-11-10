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
	viewport viewport.Model
	history  []string

	//
	handler *api.Handler

	//
	currentPrompt   *components.UserPrompt
	currentResponse *components.AgentResponse
	responseTicking bool
	eventTicking    bool
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen(h *api.Handler) tea.Model {
	hist := []string{library.Burner(0)}
	return &ChatScreen{
		viewport: viewport.New(0, 0),
		handler:  h,
		history: hist,
	}
}

func ContinueChat(h *api.Handler, from api.ContinueSessionMsg) tea.Model {
	hist := []string{library.Burner(0)}
	hist = reconstructHistory(hist, from)
	
	return &ChatScreen{
		viewport: viewport.New(0, 0),
		handler:  h,
		history: hist,
	}
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.viewport.Width, c.viewport.Height = msg.Width, msg.Height

		config.Config.SetRenderer(c.viewport.Width)
		c.history[0] = library.Burner(c.viewport.Width)

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
		if c.currentResponse != nil {
			c.currentResponse.Fail(msg.E)
			c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
		}
		c.history = append(c.history, msg.E.Error())

		return c, nil

	case api.PromptCmd:
		c.addNewPrompt(msg.Message)
		cmd := c.handler.HandlePrompt(msg.Message)

		return c, tea.Batch(cmd, api.DoTick20())

	case api.TickMsg20:
		if c.currentPrompt == nil {
			return c, nil
		}
		c.history[c.currentPrompt.HistoryIdx] = c.currentPrompt.View()

		if c.currentPrompt.Errored {
			return c, nil
		}

		if c.currentPrompt.IsSent {
			return c, nil
		}

		c.currentPrompt.Update(msg)
		return c, api.DoTick20()

	case api.TickMsg50:
		if c.currentResponse == nil {
			c.responseTicking = false
			return c, nil
		}

		c.currentResponse.Update(msg)
		c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
		if c.currentResponse.Working {
			return c, api.DoTick50()
		}
		c.responseTicking = false

		return c, nil

	case api.TickMsg60:
		if c.currentResponse == nil {
			c.eventTicking = false
			return c, nil
		}
		c.currentResponse.Update(msg)
		c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
		c.eventTicking = c.currentResponse.HasActiveTools()
		if c.eventTicking {
			return c, api.DoTick60()
		}
		return c, nil

	case api.ResponseChunkCmd:
		if c.currentPrompt != nil && !c.currentPrompt.IsSent {
			c.currentPrompt.Update(api.SentCmd{})
		}

		if c.currentResponse == nil {
			return c, nil
		}

		c.currentResponse.Update(msg)
		c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
		c.viewport.GotoBottom()
		if msg.Chunk.EOS {
			c.responseTicking = false
			return c, nil
		}
		if c.handler == nil || c.handler.Session == nil {
			return c, nil
		}
		cmds := []tea.Cmd{api.AwaitResponse(c.handler.Session.Ctx, msg.C)}
		if !c.responseTicking {
			c.responseTicking = true
			cmds = append(cmds, api.DoTick50())
		}
		return c, tea.Batch(cmds...)

	case api.EventCmd:
		switch msg.Event.(type) {
		case agent.ToolStart:
			if c.currentResponse == nil {
				return c, api.AwaitEvent(msg.Ctx, msg.C)
			}
			if c.currentPrompt != nil && !c.currentPrompt.IsSent {
				c.currentPrompt.Update(api.SentCmd{})
			}
			c.currentResponse.Update(msg.Event)
			c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
			cmds := []tea.Cmd{api.AwaitEvent(msg.Ctx, msg.C)}
			if !c.responseTicking {
				c.responseTicking = true
				cmds = append(cmds, api.DoTick50())
			}
			if !c.eventTicking {
				c.eventTicking = true
				cmds = append(cmds, api.DoTick60())
			}
			return c, tea.Batch(cmds...)

		case agent.ToolDone:
			if c.currentResponse != nil {
				c.currentResponse.Update(msg.Event)
				c.history[c.currentResponse.HistoryIdx] = c.currentResponse.View()
			}
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

func (c *ChatScreen) addNewPrompt(message string) {
	c.currentPrompt = components.NewUserPrompt(message, len(c.history))
	c.history = append(c.history, c.currentPrompt.View())
	c.currentResponse = components.NewAgentResponse(len(c.history), c.viewport.Width)
	c.history = append(c.history, c.currentResponse.View())
	c.responseTicking = false
	c.eventTicking = false
}


func reconstructHistory(h []string, from api.ContinueSessionMsg) []string {
	if from.Error == nil {
		for _, m := range from.Messages {
			h = append(h, m.Content)
		}

		return h
	}

	return append(h, from.Error.Error())
}
