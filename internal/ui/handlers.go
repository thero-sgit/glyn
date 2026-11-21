package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

func (m *Model) handleErrMsg(msg api.ErrorMsg) (tea.Model, tea.Cmd) {
	library.SendPromptBtnState = library.SendPromptBtnStates[0]
	m.modelsLoading = false
	focused := *m.focused
	updated, cmd := focused.Update(msg)
	*m.focused = updated

	return m, cmd
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	wasFocused := m.textarea.Focused()
	m.inputBoxHeight = m.inputHeight()
	m.textarea = library.CreateTextArea(msg.Width, m.inputBoxHeight-1, m.textarea.Value())
	if wasFocused {
		m.textarea.Focus()
	}
	m.textarea.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("alt+enter", "ctrl+j"),
	)

	m.updateInputLayout()
	m.updateScreenSizes()

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
			m.syncInputLayout()
			return m, nil

		case zone.Get("model-selector-btn").InBounds(msg):
			m.setFocused(&m.llmsScreen)
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

		default:
			focused := *m.focused
			updated, cmd := focused.Update(msg)
			*m.focused = updated
			return m, cmd
		}
	}

	focused := *m.focused
	updated, cmd := focused.Update(msg)
	*m.focused = updated
	return m, cmd
}

func (m *Model) handleModelsCmd(msg api.ModelsCmd) (tea.Model, tea.Cmd) {
	m.modelsLoading = false
	var cmd tea.Cmd
	m.llmsScreen, cmd = m.llmsScreen.Update(msg)
	return m, cmd
}

func (m *Model) handleModelsLoadingTick() (tea.Model, tea.Cmd) {
	if !m.modelsLoading {
		return m, nil
	}
	m.modelsSpinner, _ = m.modelsSpinner.Update(spinner.TickMsg{})
	return m, api.DoModelsLoadingTick()
}

func (m *Model) handleSessionLoadingTick() (tea.Model, tea.Cmd) {
	if !m.sessionLoading {
		return m, nil
	}
	m.sessionSpinner, _ = m.sessionSpinner.Update(spinner.TickMsg{})
	return m, api.DoSessionLoadingTick()
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

func (m *Model) handleEvent(msg api.EventCmd) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatScreen, cmd = m.chatScreen.Update(msg)
	return m, cmd
}

func (m *Model) handleGetSessionHistory(msg api.GetSessionHistory) (tea.Model, tea.Cmd) {
	m.sessionLoading = true
	return m, tea.Batch(api.DoSessionLoadingTick(), func() tea.Msg {
		return m.handler.HandleContinueSession(msg.SessionId)
	})
}

func (m *Model) handleContinueSession(msg api.ContinueSessionMsg) (tea.Model, tea.Cmd) {
	m.sessionLoading = false
	m.chatScreen = submodels.ContinueChat(&m.handler, msg)
	m.setFocused(&m.chatScreen)
	return m, nil
}

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
		m.syncInputLayout()
	} else {
		m.handleGenericKey(msg)
	}
	return m, nil
}

func (m *Model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.textarea.Focused() {
		return m.handleGenericKey(msg)
	}

	m.setFocused(&m.helpScreen)
	return m, nil
}

func (m *Model) handleSlash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() && !m.fileSelector.Focused {
		m.slashCmdCon.Focused = true
		m.syncInputLayout()
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
		var cmd, slashConCmd tea.Cmd

		slashCon := m.slashCmdCon
		slashCon.Focused = false

		switch slashCon.Options[slashCon.Cursor].Pretty() {
		case "/help":
			m.setFocused(&m.helpScreen)
		case "/sessions":
			m.setFocused(&m.sessionsScreen)
			cmd = m.handler.HandleReadSessions
		case "/model":
			m.setFocused(&m.llmsScreen)
		}

		updatedModel, slashConCmd := slashCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*components.Selector)
		m.syncInputLayout()
		return m, tea.Batch(cmd, slashConCmd)

	case m.fileSelector.Focused:
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*components.FileSelector)

		if m.fileSelector.SelectedFile != nil {
			m.textarea.InsertString(" @" + m.fileSelector.SelectedFile.Path)
			m.textarea.InsertString(fmt.Sprint(m.textarea.Length()))
			m.textarea.Focus()
			//
			m.fileSelector.Focused = false
			m.fileSelector.SelectedFile = nil
			m.fileSelector.SetOptions()
			m.syncInputLayout()
		}

		return m, cmd

	default:
		var cmd tea.Cmd
		focused := *m.focused
		focused, cmd = focused.Update(msg)

		return m, cmd
	}
}

func (m *Model) handleEscKey() (tea.Model, tea.Cmd) {
	switch {
	case m.fileSelector.Focused:
		m.fileSelector.Focused = false
		m.syncInputLayout()
		return m, nil

	case m.slashCmdCon.Focused:
		m.slashCmdCon.Focused = false
		m.syncInputLayout()
		return m, nil

	default:
		m.setFocused(&m.chatScreen)
		m.syncInputLayout()
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
	m.syncInputLayout()

	return m, tea.Batch(taCmd, VpCmd)
}

func (m *Model) sendPromptCmd() (tea.Model, tea.Cmd) {
	if strings.Trim(strings.TrimSpace(m.textarea.Value()), "\n") != "" {
		var cmd tea.Cmd

		m.chatScreen, cmd = m.chatScreen.Update(
			api.PromptCmd{
				Message: m.textarea.Value(),
			},
		)

		m.textarea.Reset()
		m.textarea.Blur()
		m.syncInputLayout()

		m.setFocused(&m.chatScreen)
		library.SendPromptBtnState = library.SendPromptBtnStates[1]
		return m, cmd
	}

	return m, nil
}

func (m *Model) syncInputLayout() {
	if m.updateInputLayout() {
		m.updateScreenSizes()
	}
}

func (m *Model) updateScreenSizes() {
	msg := tea.WindowSizeMsg{Width: m.width, Height: m.mHeight}
	for _, screen := range []*tea.Model{
		&m.chatScreen,
		&m.helpScreen,
		&m.sessionsScreen,
		&m.llmsScreen,
	} {
		updated, _ := (*screen).Update(msg)
		*screen = updated
	}
}

func (m *Model) setFocused(screen *tea.Model) {
	m.focused = screen
	if m.width == 0 || m.height == 0 {
		return
	}
	updated, _ := (*screen).Update(tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.mHeight,
	})
	*screen = updated
}
