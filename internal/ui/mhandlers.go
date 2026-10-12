package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
	"github.com/thero-sgit/xyn/internal/ui/submodels"
)

func (m *Model) handleErrMsg() (tea.Model, tea.Cmd) {
	m.chatCentre.isAgentWorking = false
	m.chatCentre.sendingPrompt = false
	stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
	m.chatCentre.currentUserPrompt.err()
	m.viewport.SetContent(strings.Join(append(m.chatCentre.prettyHistory(), m.chatCentre.currentUserPrompt.getPretty()), "\n"))

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
	m.chatCentre.updatedWidths(m.viewport.Width)

	return m, nil
}

func (m *Model) handleMouseAction(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		m.textarea.Blur()

		if m.helpScreen.Focused {
			updatedModel, cmd := m.helpScreen.Update(msg)
			m.helpScreen = *updatedModel.(*submodels.HelpModel)

			m.viewport.SetContent(m.helpScreen.View())

			return m, cmd
		}

		switch {
		case zone.Get("text-area").InBounds(msg):
			m.textarea.Focus()
			return m, nil

		case zone.Get("add-file-btn").InBounds(msg):
			m.slashCmdCon.Focused = false
			m.chatCentre.fileSelector.Focused = true
			return m, nil

		case zone.Get("model-selector-btn").InBounds(msg):
			m.slashCmdCon.Focused = false
			m.chatCentre.fileSelector.Focused = false
			return m, nil

		case zone.Get("mode-toggle-btn").InBounds(msg):
			config.Config.Mode.Toggle()
			return m, nil

		case zone.Get("send-prompt-btn").InBounds(msg):
			return m, sendPromptKeyMsg()

		case zone.Get("retry-button").InBounds(msg):
			m, cmds := CHandler.handlePrompt(m)
			return m, tea.Batch(cmds, doTick20())

		case zone.Get("thought-process-btn").InBounds(msg):
			tempPrettyHistory := m.chatCentre.prettyHistory()

			if !m.chatCentre.currentAgentRes.showReasoning {
				m.chatCentre.currentAgentRes.showReasoning = true
			} else {
				m.chatCentre.currentAgentRes.showReasoning = false
			}

			tempPrettyHistory[len(tempPrettyHistory)-1] = m.chatCentre.currentAgentRes.getPretty()
			m.viewport.SetContent(strings.Join(tempPrettyHistory, "\n"))
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) handleTick50() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.chatCentre.isAgentWorking {
		m.chatCentre.currentAgentRes.agentBgActivity.animate()
		cmd = doTick50()
	} else {
		m.renderChat()
	}

	return m, cmd
}

func (m *Model) handleTick20() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.chatCentre.sendingPrompt {
		m.viewport.SetContent(strings.Join(append(m.chatCentre.prettyHistory(), m.chatCentre.currentUserPrompt.getPretty()), "\n"))
		m.chatCentre.currentUserPrompt.animate()
		cmd = doTick20()
	}

	return m, cmd
}

func (m *Model) handleResponse(msg response) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.chatCentre.sendingPrompt {
		m.chatCentre.sendingPrompt = false
		m.chatCentre.isAgentWorking = true
		m.chatCentre.currentUserPrompt.sent()
		m.chatCentre.history = append(m.chatCentre.history, m.chatCentre.currentUserPrompt)
		cmd = doTick50()
	}

	if msg.chunk.EOS {
		m.chatCentre.isAgentWorking = false
		m.chatCentre.currentAgentRes.agentBgActivity.done(false)
		stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
		m.chatCentre.currentAgentRes.responseBuffer, _ = config.Config.MdRenderer.Render(m.chatCentre.currentAgentRes.responseBuffer)
		m.chatCentre.history = append(m.chatCentre.history, m.chatCentre.currentAgentRes)
		return m, doTick50()
	}

	if msg.chunk.Reasoning {
		if !m.chatCentre.currentAgentRes.showReasoning {
			m.chatCentre.currentAgentRes.showReasoning = true
		}

		m.chatCentre.currentAgentRes.reasoningBuffer += msg.chunk.Data
	} else {
		if m.chatCentre.currentAgentRes.showReasoning {
			m.chatCentre.currentAgentRes.showReasoning = false
		}

		if m.chatCentre.isAgentWorking {
			m.chatCentre.currentAgentRes.agentBgActivity.done(false)
		}
		m.chatCentre.currentAgentRes.responseBuffer += msg.chunk.Data
	}

	m.viewport.SetContent(strings.Join(append(m.chatCentre.prettyHistory(), m.chatCentre.currentAgentRes.getPretty()), "\n"))

	return m, tea.Batch(awaitResponse(CHandler.session.Ctx, msg.c), cmd)
}

func (m *Model) handleSessionInfo(msg sessionInfo) (tea.Model, tea.Cmd) {
	m.chatCentre.sessionName += msg.chunk.Data
	if msg.chunk.EOS {
		return m, nil
	}

	return m, awaitSessionInfo(CHandler.session.Ctx, msg.c)
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
	case "pgup", "pgdown", "up", "down":
		return m.handleUpDownArrows(msg)
	case "left", "right":
		return m.handleLeftRightArrows(msg)
	case "/":
		return m.handleSlash(msg)
	case "enter":
		return m.handleEnterKey(msg)
	case "?":
		return m.handleHelpKey()
	case "tab":
		return m.handleTabKey(msg)
	case "@":
		return m.handleAtKey()
	}

	return m.handleGenericKey(msg)
}

func (m *Model) handleUpDownArrows(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var VpCmd, taCmd, fVpCmd, slConCmd tea.Cmd
	if !m.textarea.Focused() {
		m.viewport, VpCmd = m.viewport.Update(msg)
	} else {
		m.textarea, taCmd = m.textarea.Update(msg)
	}

	if m.chatCentre.fileSelector.Focused {
		updatedModel, cmd := m.chatCentre.fileSelector.Update(msg)
		m.chatCentre.fileSelector = *updatedModel.(*submodels.FileSelector)

		return m, cmd
	}

	if m.slashCmdCon.Focused {
		updatedModel, cmd := m.slashCmdCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*submodels.Selector)

		return m, cmd
	}

	return m, tea.Batch(VpCmd, taCmd, fVpCmd, slConCmd)
}

func (m *Model) handleAtKey() (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() && !m.slashCmdCon.Focused {
		m.chatCentre.fileSelector.Focused = true
	}
	return m, nil
}

func (m *Model) handleHelpKey() (tea.Model, tea.Cmd) {
	m.helpScreen.Focused = true
	m.viewport.SetContent(m.helpScreen.View())

	return m, nil
}

func (m *Model) handleSlash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.textarea.Focused() && !m.chatCentre.fileSelector.Focused {
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
	c := m.chatCentre
	if c.fileSelector.Focused && len(c.fileSelector.Options) > 0 {
		updatedModel, cmd := m.chatCentre.fileSelector.Update(msg)
		m.chatCentre.fileSelector = *updatedModel.(*submodels.FileSelector)

		return m, cmd
	}

	m.chatCentre = c
	return m, nil
}

func (m *Model) handleEnterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.chatCentre.fileSelector.Focused {
		if len(m.chatCentre.fileSelector.Options) > 0 {
			file := m.chatCentre.fileSelector.Options[m.chatCentre.fileSelector.Cursor]
			if m.chatCentre.fileSelector.Prompt == "rm " {
				m.chatCentre.removeFileForContext(file.Path)
				m.chatCentre.fileSelector.Focused = false

				m.chatCentre.fileSelector.SetPrompt("@")
				return m, nil
			}

			if !file.IsDir {
				m.chatCentre.addFileForContext(file.Path)
				m.chatCentre.fileSelector.Focused = false

				m.chatCentre.fileSelector.SetPrompt("@")
				return m, nil
			}			
		}
	}

	if m.slashCmdCon.Focused {
		slashCon := m.slashCmdCon
		slashCon.Focused = false

		if slashCon.Options[slashCon.Cursor].Name() == "/help" {
			m.handleHelpKey()
		}

		if slashCon.Options[slashCon.Cursor].Name() == "/sessions" {
			m.sessionsScreen.Focused = true
			m.viewport.SetContent(m.sessionsScreen.View())
		}

		updatedModel, cmd := slashCon.Update(msg)
		m.slashCmdCon = *updatedModel.(*submodels.Selector)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleLeftRightArrows(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.helpScreen.Focused {
		updatedModel, cmd := m.helpScreen.Update(msg)
		m.helpScreen = *updatedModel.(*submodels.HelpModel)

		m.viewport.SetContent(m.helpScreen.View())

		return m, cmd
	}

	return m, nil
}

func (m *Model) handleAltEnter() (tea.Model, tea.Cmd) {
	if m.chatCentre.isAgentWorking {
		return m, nil
	}

	prompt := strings.Trim(strings.TrimSpace(m.textarea.Value()), "\n")

	var animateCmd tea.Cmd
	var cmds tea.Cmd

	if prompt != "" {
		m.chatCentre.currentUserPrompt = newUserPrompt(prompt, m.viewport.Width)
		m.chatCentre.currentAgentRes = newAgentResponse(m.viewport.Width)
		m.chatCentre.currentAgentRes.index = max(0, len(m.chatCentre.history)-1)
		m, cmds = CHandler.handlePrompt(m)
		m.viewport.GotoBottom()

		animateCmd = doTick20()
	}

	return m, tea.Batch(cmds, animateCmd)
}

func (m *Model) handleEscKey() (tea.Model, tea.Cmd) {
	if m.chatCentre.fileSelector.Focused {
		m.chatCentre.fileSelector.Focused = false
		return m, nil
	}

	if m.helpScreen.Focused {
		m.helpScreen.Focused = false
		m.renderChat()
		return m, nil
	}

	if m.sessionsScreen.Focused {
		m.sessionsScreen.Focused = false
		m.renderChat()
		return m, nil
	}

	if m.slashCmdCon.Focused {
		m.slashCmdCon.Focused = false
		return m, nil
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

	if m.chatCentre.fileSelector.Focused {
		updatedModel, cmd := m.chatCentre.fileSelector.Update(msg)
		m.chatCentre.fileSelector = *updatedModel.(*submodels.FileSelector)

		return m, cmd
	}

	// if m.chatCentre.fileInpBoxOpen {
	// 	Ti := m.chatCentre.fileInpBox.Ti
	// 	Vp := m.chatCentre.fileInpBox.Vp

	// 	if Ti.Value() == "rm " {
	// 		Ti = m.chatCentre.fileInpBox.SetPrompt(Ti, "rm ")
	// 		Ti.Reset()
	// 	}

	// 	matches, err := common.ListMatches(Ti.Value())

	// 	if err != nil {
	// 		Vp.GotoTop()
	// 		Vp.SetContent(err.Error())
	// 	} else {
	// 		listItemStyle := lipgloss.NewStyle().MarginLeft(2)
	// 		var mList []string
	// 		for _, e := range matches {
	// 			if e.IsDir {
	// 				mList = append(mList, listItemStyle.Bold(true).Foreground(stylesandcomponents.AccentColor).Render(e.Name))
	// 			} else {
	// 				mList = append(mList, listItemStyle.Render(e.Name))
	// 			}
	// 		}
	// 		Vp.SetContent(strings.Join(mList, "\n"))
	// 	}

	// 	m.chatCentre.fileInpBox.Ti = Ti
	// 	m.chatCentre.fileInpBox.Vp = Vp
	// 	m.chatCentre.fileInpBox.Matches = matches
	// }

	if m.chatCentre.sendingPrompt {
		stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[1]
	} else {
		stylesandcomponents.SendPromptState = stylesandcomponents.SendPromptBtnStates[0]
	}

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, VpCmd = m.viewport.Update(msg)

	return m, tea.Batch(taCmd, VpCmd)
}
