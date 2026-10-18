package ui

import (
	// "strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
	"github.com/thero-sgit/xyn/internal/ui/submodels"
)

func (m *Model) handleErrMsg() (tea.Model, tea.Cmd) {
	// m.isAgentWorking = false
	// m.sendingPrompt = false
	// stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
	// m.currentUserPrompt.err()
	// m.viewport.SetContent(strings.Join(append(m.prettyHistory(), m.currentUserPrompt.getPretty()), "\n"))

	return m, nil
}

func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	taHeight := 2
	if m.textarea.Length() > 0 {
		taHeight = 6
	}

	m.textarea = stylesandcomponents.CreateTextArea(msg.Width, taHeight, m.textarea.Value())

	headerHeight := 2
	footerHeight := 2
	middleHeight := m.height - headerHeight - footerHeight
	promptBoxHeight := taHeight + 1
	m.viewport.Height = max(middleHeight-promptBoxHeight, 1)
	m.viewport.Width = m.width - 4

	config.Config.SetRenderer(m.viewport.Width - 2)
	// m.updatedWidths(m.viewport.Width)
	// m.sessionsScreen.Width = m.viewport.Width

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
			// m.slashCmdCon.Focused = false
			// m.fileSelector.Focused = true
			return m, nil

		case zone.Get("model-selector-btn").InBounds(msg):
			// m.slashCmdCon.Focused = false
			// m.helpScreen.Focused = false
			// m.sessionsScreen.Focused = false
			// m.llmsScreen.Focused = true
			m.viewport.SetContent(m.llmsScreen.View())
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
			return m, sendPromptKeyMsg()

		case zone.Get("retry-button").InBounds(msg):
			m, cmds := CHandler.handlePrompt(m)
			return m, tea.Batch(cmds, doTick20())

		// case zone.Get("thought-process-btn").InBounds(msg):
		// 	tempPrettyHistory := m.prettyHistory()

		// 	if !m.currentAgentRes.showReasoning {
		// 		m.currentAgentRes.showReasoning = true
		// 	} else {
		// 		m.currentAgentRes.showReasoning = false
		// 	}

		// 	tempPrettyHistory[len(tempPrettyHistory)-1] = m.currentAgentRes.getPretty()
		// 	m.viewport.SetContent(strings.Join(tempPrettyHistory, "\n"))
		// 	return m, nil

		// case m.helpScreen.Focused:
		// 	updatedModel, cmd := m.helpScreen.Update(msg)
		// 	m.helpScreen = *updatedModel.(*submodels.HelpModel)

		// 	m.viewport.SetContent(m.helpScreen.View())

		// 	return m, cmd

		// case m.llmsScreen.Focused:
		// 	updatedModel, cmd := m.llmsScreen.Update(msg)
		// 	m.llmsScreen = *updatedModel.(*submodels.LLModels)

		// 	m.viewport.SetContent(m.llmsScreen.View())

		// 	return m, cmd
		}
	}

	
	// if !(m.helpScreen.Focused && m.sessionsScreen.Focused && m.llmsScreen.Focused) {
	// 	var cmd tea.Cmd
	// 	m.viewport, cmd = m.viewport.Update(msg)
	// 	return m, cmd
	// }	

	return m, nil
}

func (m *Model) handleTick50() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// if m.isAgentWorking {
	// 	m.currentAgentRes.agentBgActivity.animate()
	// 	cmd = doTick50()
	// } else {
	// 	// m.renderChat()
	// }

	return m, cmd
}

func (m *Model) handleTick20() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// if m.sendingPrompt {
	// 	m.viewport.SetContent(strings.Join(append(m.prettyHistory(), m.currentUserPrompt.getPretty()), "\n"))
	// 	m.currentUserPrompt.animate()
	// 	cmd = doTick20()
	// }

	return m, cmd
}

func (m *Model) handleResponse(msg response) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// if m.sendingPrompt {
	// 	m.sendingPrompt = false
	// 	m.isAgentWorking = true
	// 	m.currentUserPrompt.sent()
	// 	m.history = append(m.history, m.currentUserPrompt)
	// 	cmd = doTick50()
	// }

	// if msg.chunk.EOS {
	// 	m.isAgentWorking = false
	// 	m.currentAgentRes.agentBgActivity.done(false)
	// 	stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
	// 	m.currentAgentRes.responseBuffer, _ = config.Config.MdRenderer.Render(m.currentAgentRes.responseBuffer)
	// 	m.history = append(m.history, m.currentAgentRes)
	// 	return m, doTick50()
	// }

	// if msg.chunk.Reasoning {
	// 	if !m.currentAgentRes.showReasoning {
	// 		m.currentAgentRes.showReasoning = true
	// 	}

	// 	m.currentAgentRes.reasoningBuffer += msg.chunk.Data
	// } else {
	// 	if m.currentAgentRes.showReasoning {
	// 		m.currentAgentRes.showReasoning = false
	// 	}

	// 	if m.isAgentWorking {
	// 		m.currentAgentRes.agentBgActivity.done(false)
	// 	}
	// 	m.currentAgentRes.responseBuffer += msg.chunk.Data
	// }

	// m.viewport.SetContent(strings.Join(append(m.prettyHistory(), m.currentAgentRes.getPretty()), "\n"))

	return m, tea.Batch(awaitResponse(CHandler.session.Ctx, msg.c), cmd)
}

// func (m *Model) handleSessionInfo(msg sessionInfo) (tea.Model, tea.Cmd) {
// 	m.sessionName += msg.chunk.Data
// 	if msg.chunk.EOS {
// 		return m, nil
// 	}

// 	return m, awaitSessionInfo(CHandler.session.Ctx, msg.c)
// }

func (m *Model) handleModelsCmd(msg submodels.ModelsCmd) (tea.Model, tea.Cmd) {
	// updatedModel, cmd := m.llmsScreen.Update(msg)
	// // m.llmsScreen = *updatedModel.(*submodels.LLModels)
	 return m, nil
}

// KEYS
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		CHandler.session.Close()
		return m, tea.Quit

	case "esc":
		return m.handleEscKey()
	case "ctrl+e":
		return m.handleCtrlE()
	case "alt+enter":
		return m.handleAltEnter()
	case "pgup", "pgdown", "up", "down", "left", "right":
		return m.handleUpArrows(msg)
	case "/":
		return m.handleSlash(msg)
	case "enter":
		return m.handleEnterKey(msg)
	case "?":
		return m.handleHelpKey(msg)
	case "tab":
		return m.handleTabKey(msg)
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
		m.slashCmdCon = *updatedModel.(*submodels.Selector)

		return m, cmd

	case m.fileSelector.Focused:
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*submodels.FileSelector)

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
		m.handleGenericKey(msg)

	} else {
		// m.helpScreen.Focused = true
		// m.sessionsScreen.Focused = false
		// m.llmsScreen.Focused = false
		// m.viewport.SetContent(m.helpScreen.View())
	}

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

func (m *Model) handleTabKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.fileSelector.Focused && len(m.fileSelector.Options) > 0 {
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*submodels.FileSelector)

		return m, cmd
	}

	// if m.slashCmdCon.Focused {
	// 	updatedModel, cmd := m.slashCmdCon.Update(msg)
	// 	m.slashCmdCon = *updatedModel.(*submodels.Selector)

	// 	return m, cmd
	// }

	return m, nil
}

func (m *Model) handleEnterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// case m.fileSelector.Focused:
	// 	if len(m.fileSelector.Options) > 0 {
	// 		file := m.fileSelector.Options[m.fileSelector.Cursor]
	// 		if m.fileSelector.Prompt == "rm " {
	// 			m.removeFileForContext(file.Path)
	// 			m.fileSelector.Focused = false

	// 			m.fileSelector.SetPrompt("@")
	// 			return m, nil
	// 		}

	// 		if !file.IsDir {
	// 			m.addFileForContext(file.Path)
	// 			m.fileSelector.Focused = false

	// 			m.fileSelector.SetPrompt("@")
	// 			return m, nil
	// 		} else {
	// 			updatedModel, cmd := m.fileSelector.Update(msg)
	// 			m.fileSelector = *updatedModel.(*submodels.FileSelector)

	// 			return m, cmd
	// 		}
	// 	}

	// case m.slashCmdCon.Focused:
	// 	slashCon := m.slashCmdCon
	// 	slashCon.Focused = false

	// 	if slashCon.Options[slashCon.Cursor].Pretty() == "/help" {
	// 		m.handleHelpKey(msg)
	// 	}

	// 	if slashCon.Options[slashCon.Cursor].Pretty() == "/sessions" {
	// 		m.helpScreen.Focused = false
	// 		m.sessionsScreen.Focused = true
	// 		m.llmsScreen.Focused = false
	// 		m.viewport.SetContent(m.sessionsScreen.View())
	// 	}

	// 	if slashCon.Options[slashCon.Cursor].Pretty() == "/model" {
	// 		m.helpScreen.Focused = false
	// 		m.sessionsScreen.Focused = false
	// 		m.llmsScreen.Focused = true
	// 		m.viewport.SetContent(m.llmsScreen.View())
	// 	}

	// 	updatedModel, cmd := slashCon.Update(msg)
	// 	m.slashCmdCon = *updatedModel.(*submodels.Selector)
	// 	return m, cmd
	
	// case m.llmsScreen.Focused:
	// 	updatedModel, cmd := m.llmsScreen.Update(msg)
	// 	m.llmsScreen = *updatedModel.(*submodels.LLModels)
	// 	m.viewport.SetContent(m.llmsScreen.View())
	// 	return m, cmd

	}

	return m, nil
}

func (m *Model) handleLeftRightArrows(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	focused := *m.focused
	focused.Update(msg)

	return m, nil
}

func (m *Model) handleAltEnter() (tea.Model, tea.Cmd) {
	// if m.isAgentWorking {
	// 	return m, nil
	// }

	// prompt := strings.Trim(strings.TrimSpace(m.textarea.Value()), "\n")

	var animateCmd tea.Cmd
	var cmds tea.Cmd

	// if prompt != "" {
	// 	m.currentUserPrompt = newUserPrompt(prompt, m.viewport.Width)
	// 	m.currentAgentRes = newAgentResponse(m.viewport.Width)
	// 	m.currentAgentRes.index = max(0, len(m.history)-1)
	// 	m, cmds = CHandler.handlePrompt(m)
	// 	m.viewport.GotoBottom()

	// 	animateCmd = doTick20()
	// }

	return m, tea.Batch(cmds, animateCmd)
}

func (m *Model) handleEscKey() (tea.Model, tea.Cmd) {
	switch {
	// case m.fileSelector.Focused:
	// 	m.fileSelector.Focused = false
	// 	return m, nil

	// case m.helpScreen.Focused:
	// 	m.helpScreen.Focused = false
	// 	m.renderChat()
	// 	return m, nil

	// case m.sessionsScreen.Focused:
	// 	m.sessionsScreen.Focused = false
	// 	m.renderChat()
	// 	return m, nil

	// case m.llmsScreen.Focused:
	// 	m.llmsScreen.Focused = false
	// 	m.renderChat()
	// 	return m, nil

	// case m.slashCmdCon.Focused:
	// 	m.slashCmdCon.Focused = false
	// 	return m, nil
	}

	CHandler.session.Close()
	return m, tea.Quit
}

func (m *Model) handleGenericKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var taCmd, VpCmd tea.Cmd
	if m.slashCmdCon.Focused {
		updatedModel, cmd := m.slashCmdCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*submodels.Selector)

		return m, cmd
	}

	if m.fileSelector.Focused {
		updatedModel, cmd := m.fileSelector.Update(msg)
		m.fileSelector = *updatedModel.(*submodels.FileSelector)

		return m, cmd
	}

	// if m.sendingPrompt {
	// 	stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[1]
	// } else {
	// 	stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
	// }

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, VpCmd = m.viewport.Update(msg)

	return m, tea.Batch(taCmd, VpCmd)
}
