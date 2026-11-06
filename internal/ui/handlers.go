package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

func (m *Model) handleErrMsg(msg api.ErrorMsg) (tea.Model, tea.Cmd) {
	library.SendPromptBtnState = library.SendPromptBtnStates[0]
	focused := *m.focused
	focused.Update(msg)

	return m, nil
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	taHeight := 2
	if m.textarea.Length() > 0 {
		taHeight = 6
	}

	wasFocused := m.textarea.Focused()
	m.textarea = library.CreateTextArea(msg.Width, taHeight, m.textarea.Value())
	if wasFocused {
		m.textarea.Focus()
	}
	m.textarea.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("alt+enter", "ctrl+j"),
	)

	return m, nil
}

func (m *Model) handleMouseAction(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		m.textarea.Blur()

		switch {
		case zone.Get("text-area").InBounds(msg):
			m.textarea.Focus()
			return m, nil

		case zone.Get("add-file-btn").InBounds(msg):
			m.slashCmdCon.Focused = false
			m.fileSelector.Focused = true
			return m, nil

		case zone.Get("model-selector-btn").InBounds(msg):
			m.focused = &m.llmsScreen
			return m, nil

		case zone.Get("mode-toggle-btn").InBounds(msg):
			config.Config.Mode.Toggle()
			switch config.Config.Mode.Name {
			case "ask ":
				config.Config.Mode.Icon = " ?"
			case "agent":
				config.Config.Mode.Icon = "\u2726"
			}
			return m, nil

		case zone.Get("send-prompt-btn").InBounds(msg):
			return m.sendPromptCmd()

		default:
			focused := *m.focused
			focused.Update(msg)
			return m, nil
		}
	}

	focused := *m.focused
	focused.Update(msg)

	return m, nil
}

func (m *Model) handleModelsCmd(msg api.ModelsCmd) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.llmsScreen, cmd = m.llmsScreen.Update(msg)
	return m, cmd
}

func (m *Model) handleTick20(msg api.TickMsg20) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)

	return m, cmd
}

func (m *Model) handleTick50(msg api.TickMsg50) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)

	return m, cmd
}

func (m *Model) handleTick60(msg api.TickMsg60) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)

	return m, cmd
}

func (m *Model) handleResponseChunkCmd(msg api.ResponseChunkCmd) (tea.Model, tea.Cmd) {
	if msg.Chunk.EOS {
		library.SendPromptBtnState = library.SendPromptBtnStates[0]
	}

	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)

	return m, cmd
}

func (m *Model) handleEvent(msg api.EventMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)
	return m, cmd
}

// func (m *Model) handleSessionInfo(msg sessionInfo) (tea.Model, tea.Cmd) {
// 	m.sessionName += msg.chunk.Data
// 	if msg.chunk.EOS {
// 		return m, nil
// 	}

// 	return m, awaitSessionInfo(CHandler.session.Ctx, msg.c)
// }


// KEYS
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		return m.handleEscKey()
	case "ctrl+e":
		return m.handleCtrlE()
	case "ctrl+n":
		return m.handleCtrlEnter(msg)
	case "pgup", "pgdown", "up", "down", "left", "right", "tab":
		return m.handleUpArrows(msg)
	case "/":
		return m.handleSlash(msg)
	case "enter":
		return m.handleEnterKey(msg)
	case "?":
		return m.handleHelpKey(msg)
	case "@":
		return m.handleAtKey(msg)
	}

	return m.handleGenericKey(msg)
}

func (m *Model) handleUpArrows(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.textarea.Focused():
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd 

	case m.slashCmdCon.Focused:
		updatedModel, cmd := m.slashCmdCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*components.Selector)

		return m, cmd

	case m.fileSelector.Focused:
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*components.FileSelector)

		return m, cmd

	default:
		focused := *m.focused
		focused.Update(msg)

		return m, nil
	}	
}

func (m *Model) handleAtKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() && !m.slashCmdCon.Focused {
		m.fileSelector.Focused = true
	} else {
		m.handleGenericKey(msg)
	}
	return m, nil
}

func (m *Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.textarea.Focused() {
		return m.handleGenericKey(msg)
	}

	m.focused = &m.helpScreen
	return m, nil
}

func (m *Model) handleSlash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() && !m.fileSelector.Focused {
		m.slashCmdCon.Focused = true
	} else {
		m.handleGenericKey(msg)
	}
	return m, nil
}

func (m *Model) handleCtrlE() (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() {
		m.textarea.Focus()
	} else {
		m.textarea.Blur()
	}
	return m, nil
}

func (m *Model) handleCtrlEnter(msg tea.Msg) (tea.Model, tea.Cmd) {
	focused := *m.focused
	focused.Update(msg)
	return m, nil
}

func (m *Model) handleEnterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.textarea.Focused():
		return m.sendPromptCmd()

	case m.slashCmdCon.Focused:
		slashCon := m.slashCmdCon
		slashCon.Focused = false

		switch slashCon.Options[slashCon.Cursor].Pretty() {		
		case "/help":
			m.focused = &m.helpScreen
		case"/sessions":
			m.focused = &m.sessionsScreen
		case"/model":
			m.focused = &m.llmsScreen
		}

		updatedModel, cmd := slashCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*components.Selector)
		return m, cmd

	default:
		focused := *m.focused
		focused.Update(msg)

		return m, nil
	}
}

func (m *Model) handleEscKey() (tea.Model, tea.Cmd) {
	switch {
	case m.fileSelector.Focused:
		m.fileSelector.Focused = false
		return m, nil

	case m.slashCmdCon.Focused:
		m.slashCmdCon.Focused = false
		return m, nil

	default:
		m.focused = &m.chatScreen
		return m, nil
	}
}

func (m *Model) handleGenericKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var taCmd, VpCmd tea.Cmd
	if m.slashCmdCon.Focused {
		updatedModel, cmd := m.slashCmdCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*components.Selector)

		return m, cmd
	}

	if m.fileSelector.Focused {
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*components.FileSelector)

		return m, cmd
	}

	if !m.textarea.Focused() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok &&
			(keyMsg.Type == tea.KeyRunes || keyMsg.Type == tea.KeySpace) {
			m.textarea.Focus()
		}
	}

	m.textarea, taCmd = m.textarea.Update(msg)

	return m, tea.Batch(taCmd, VpCmd)
}

func (m *Model) sendPromptCmd() (tea.Model, tea.Cmd) {
	if strings.Trim(strings.TrimSpace(m.textarea.Value()), "\n") != "" {		
		var cmd tea.Cmd

		m.chatScreen, cmd = m.chatScreen.Update(
			api.PromptCmd {
				Message: m.textarea.Value(),
			},
		)

		m.textarea.Reset()
		m.textarea.Blur()

		m.focused = &m.chatScreen
		library.SendPromptBtnState = library.SendPromptBtnStates[1]
		return m, cmd
	}

	return m, nil
}
