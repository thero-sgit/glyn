package submodels

import (
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ai/common"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

type ChatScreen struct {
	viewport     viewport.Model
	emptyView    string
	errorView    string
	turns        []chatTurn
	pageStart    int
	contentDirty bool
	cachedView   string

	//
	handler *api.Handler

	//
	currentPrompt   *components.UserPrompt
	currentRun      *components.Run
	responseTicking bool
	eventTicking    bool
}

type chatTurn struct {
	prompt   *components.UserPrompt
	run      *components.Run
	wrappers []common.MessageWrapper
}

func (c *ChatScreen) Init() tea.Cmd {
	return nil
}

func NewChatScreen(h *api.Handler) tea.Model {
	return &ChatScreen{
		viewport:     viewport.New(0, 0),
		handler:      h,
		emptyView:    library.Burner(0),
		contentDirty: true,
	}
}

func ContinueChat(h *api.Handler, from api.ContinueSessionMsg) tea.Model {
	chat := &ChatScreen{
		viewport:     viewport.New(0, 0),
		handler:      h,
		emptyView:    library.Burner(0),
		contentDirty: true,
	}
	chat.reconstructTurns(from)
	return chat
}

func (c *ChatScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if c.viewport.Width != msg.Width || c.viewport.Height != msg.Height {
			c.contentDirty = true
		}
		widthChanged := c.viewport.Width != msg.Width
		c.viewport.Width, c.viewport.Height = msg.Width, msg.Height

		if widthChanged {
			config.Config.SetRenderer(c.viewport.Width)
		}
		c.emptyView = library.Burner(c.viewport.Width)
		for i := range c.turns {
			if c.turns[i].run != nil {
				c.turns[i].run.SetWidth(c.viewport.Width)
			}
		}

		return c, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		c.contentDirty = true
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			pageEnd := min(c.pageStart+1, len(c.turns))
			var cmds []tea.Cmd
			for index := c.pageStart; index < pageEnd; index++ {
				turn := &c.turns[index]
				if turn.prompt != nil {
					if message, retry := turn.prompt.RetryClicked(msg); retry {
						return c, c.retryTurn(index, message)
					}
				}
				if turn.run == nil {
					continue
				}
				_, runCmd := turn.run.Update(msg)
				if runCmd != nil {
					cmds = append(cmds, runCmd)
				}
			}
			cmd = tea.Batch(cmds...)
			return c, cmd
		}

		c.viewport, cmd = c.viewport.Update(msg)
		return c, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "pgup":
			if c.pageStart > 0 {
				c.pageStart--
				c.contentDirty = true
				c.viewport.GotoTop()
			}
		case "down", "pgdown":
			if c.pageStart+1 < len(c.turns) {
				c.pageStart++
				c.contentDirty = true
				c.viewport.GotoTop()
			}
		}

	case api.ErrorMsg:
		c.contentDirty = true
		if c.currentPrompt != nil {
			c.currentPrompt.Update(msg)
		}
		if c.currentRun != nil {
			c.currentRun.Fail(msg.E)
		}
		c.errorView = msg.E.Error()

		return c, nil

	case api.PromptCmd:
		c.contentDirty = true
		c.addNewPrompt(msg.Message)
		cmd := c.handler.HandlePrompt(msg.Message)

		return c, tea.Batch(cmd, api.DoTick20())

	case api.TickMsg20:
		if c.currentPrompt == nil {
			return c, nil
		}

		if c.currentPrompt.Errored {
			return c, nil
		}

		if c.currentPrompt.IsSent {
			return c, nil
		}

		c.contentDirty = true
		c.currentPrompt.Update(msg)
		return c, api.DoTick20()

	case api.TickMsg50:
		if c.currentRun == nil {
			c.responseTicking = false
			return c, nil
		}
		c.contentDirty = true
		c.currentRun.Update(msg)
		if c.currentRun.Working {
			return c, api.DoTick50()
		}
		c.responseTicking = false

		return c, nil

	case api.TickMsg60:
		if c.currentRun == nil {
			c.eventTicking = false
			return c, nil
		}
		c.contentDirty = true
		c.currentRun.Update(msg)
		c.eventTicking = c.currentRun.HasActiveTools()
		if c.eventTicking {
			return c, api.DoTick60()
		}
		return c, nil

	case api.ResponseChunkCmd:
		c.contentDirty = true
		if c.currentPrompt != nil && !c.currentPrompt.IsSent {
			c.currentPrompt.Update(api.SentCmd{})
		}

		if c.currentRun == nil {
			return c, nil
		}

		c.currentRun.Update(msg)
		if msg.Chunk.EOS {
			c.errorView = ""
		}
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
		case common.ToolStart:
			if c.currentRun == nil {
				return c, api.AwaitEvent(msg.Ctx, msg.C)
			}
			if c.currentPrompt != nil && !c.currentPrompt.IsSent {
				c.currentPrompt.Update(api.SentCmd{})
			}
			c.contentDirty = true
			c.currentRun.Update(msg.Event)
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

		case common.ToolDone:
			if c.currentRun != nil {
				c.contentDirty = true
				c.currentRun.Update(msg.Event)
			}
			return c, api.AwaitEvent(msg.Ctx, msg.C)
		}
	}

	return c, nil
}

func (c *ChatScreen) View() string {
	if !c.contentDirty {
		return c.cachedView
	}

	var content string
	switch {
	case len(c.turns) > 0:
		var sections []string
		pageEnd := min(c.pageStart+1, len(c.turns))
		if c.pageStart > 0 || pageEnd < len(c.turns) {
			sections = append(sections, c.turnNavigationHint())
		}
		for _, turn := range c.turns[c.pageStart:pageEnd] {
			sections = append(sections, turn.prompt.View(), turn.run.View())
		}
		content = lipgloss.JoinVertical(lipgloss.Top, sections...)
		if c.errorView != "" && len(c.turns)-1 < pageEnd {
			content = lipgloss.JoinVertical(lipgloss.Top, content, c.errorView)
		}
	case c.errorView != "":
		content = c.errorView
	default:
		content = lipgloss.NewStyle().
			Width(c.viewport.Width).
			Height(c.viewport.Height).
			AlignVertical(lipgloss.Center).
			Render(c.emptyView)
	}

	c.viewport.SetContent(content)
	c.cachedView = c.viewport.View()
	c.contentDirty = false
	return c.cachedView
}

func (c *ChatScreen) turnNavigationHint() string {
	var hints []string
	if c.pageStart > 0 {
		hints = append(hints, lipgloss.NewStyle().
			Foreground(library.AccentColor).
			Bold(true).
			Render("↑ older"))
	}
	if c.pageStart+1 < len(c.turns) {
		hints = append(hints, lipgloss.NewStyle().
			Foreground(library.AccentColor).
			Bold(true).
			Render("↓ newer"))
	}
	hints = append(hints, library.SubtleStyle.Render("·  up/down or pgup/pgdn"))
	return lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		PaddingLeft(1).
		PaddingRight(1).
		Render(strings.Join(hints, "    "))
}

func (c *ChatScreen) addNewPrompt(message string) {
	c.contentDirty = true
	c.errorView = ""
	index := len(c.turns) * 2
	c.currentPrompt = components.NewUserPrompt(message, index)
	c.currentRun = components.NewRun("", index+1, c.viewport.Width, message)
	c.turns = append(c.turns, chatTurn{prompt: c.currentPrompt, run: c.currentRun})
	c.pageStart = len(c.turns) - 1
	c.responseTicking = false
	c.eventTicking = false
}

func (c *ChatScreen) retryTurn(index int, message string) tea.Cmd {
	c.contentDirty = true
	if index == len(c.turns)-1 {
		prompt := components.NewUserPrompt(message, index*2)
		run := components.NewRun("", index*2+1, c.viewport.Width, message)
		c.turns[index] = chatTurn{prompt: prompt, run: run}
		c.currentPrompt = prompt
		c.currentRun = run
		c.pageStart = index
		c.errorView = ""
		c.responseTicking = false
		c.eventTicking = false
	} else {
		c.addNewPrompt(message)
	}
	c.viewport.GotoTop()

	return tea.Batch(c.handler.HandlePrompt(message), api.DoTick20())
}

func (c *ChatScreen) reconstructTurns(from api.ContinueSessionMsg) {
	if from.Error != nil {
		c.errorView = from.Error.Error()
	}

	wrappers := append([]common.MessageWrapper(nil), from.Messages...)
	sort.SliceStable(wrappers, func(i, j int) bool {
		return wrappers[i].Time < wrappers[j].Time
	})

	runIndexes := make(map[string]int)
	var currentRun int
	hasCurrentRun := false
	for _, wrapper := range wrappers {
		runIndex, exists := runIndexes[wrapper.ID]
		if wrapper.ID == "" {
			if wrapper.Msg.Role == "user" || !hasCurrentRun {
				runIndex = len(c.turns)
				c.turns = append(c.turns, chatTurn{})
				currentRun = runIndex
				hasCurrentRun = true
			} else {
				runIndex = currentRun
			}
		} else if !exists {
			runIndex = len(c.turns)
			runIndexes[wrapper.ID] = runIndex
			c.turns = append(c.turns, chatTurn{})
		}

		turn := &c.turns[runIndex]
		turn.wrappers = append(turn.wrappers, wrapper)
		if wrapper.Msg.Role == "user" {
			turn.prompt = components.NewUserPrompt(wrapper.Msg.Content, runIndex*2)
			turn.prompt.Update(api.SentCmd{})
		}
	}

	for runIndex := range c.turns {
		turn := &c.turns[runIndex]
		if turn.prompt == nil {
			continue
		}
		turn.run = components.NewHistoricalRun(turn.wrappers[0].ID, runIndex*2+1, c.viewport.Width, turn.wrappers)
	}
	c.contentDirty = true

	if len(c.turns) > 0 {
		last := c.turns[len(c.turns)-1]
		c.currentPrompt = last.prompt
		c.currentRun = last.run
		c.pageStart = len(c.turns) - 1
	}
}
