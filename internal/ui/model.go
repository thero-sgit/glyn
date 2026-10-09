package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/utils"
)

type tickMsg50 time.Time

func doTick50() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg50(t)
	})
}

type tickMsg20 time.Time

func doTick20() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg20(t)
	})
}

func sendPromptKeyMsg() tea.Cmd {
	return func() tea.Msg {
		return tea.KeyMsg{
			Type: tea.KeyEnter,
			Alt:  true,
		}
	}
}

type Model struct {
	width    int
	height   int
	textarea textarea.Model
	viewport viewport.Model

	//
	chatCentre chatCentre

	//
	slashCmdCon slashCmdCtrl
}

func InitialModel() *Model {
	vp := viewport.New(0, 0)

	vp.KeyMap = viewport.KeyMap{
		Up:           key.NewBinding(key.WithKeys("up")),
		Down:         key.NewBinding(key.WithKeys("down")),
		PageUp:       key.NewBinding(key.WithKeys("pgup")),
		PageDown:     key.NewBinding(key.WithKeys("pgdn")),
		HalfPageUp:   key.NewBinding(),
		HalfPageDown: key.NewBinding(),
	}

	model := Model{
		chatCentre:  newChatCentre(),
		slashCmdCon: newSlashCmdCtrl(0, 0),
		viewport:    vp,
	}

	model.viewport.SetContent(strings.Join(model.chatCentre.prettyHistory(), "\n"))

	InitHandler()

	return &model
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd    tea.Cmd
		vpCmd  tea.Cmd
		taCmd  tea.Cmd
		ciCmd  tea.Cmd
		mniCmd tea.Cmd
		fvpCmd tea.Cmd
		ftiCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.MouseMsg:
		return m.handleMouseAction(msg)
	case errMsg:
		return m.handleErrMsg(msg)
	case tickMsg50:
		return m.handleTick50()
	case tickMsg20:
		return m.handleTick20()
	case response:
		return m.handleResponse(msg)
	case sessionInfo:
		return m.handleSessionInfo(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			CHandler.session.Close()
			return m, tea.Quit

		case "esc":
			if m.slashCmdCon.acceptingCmd || m.slashCmdCon.openCmd != "" {
				m.slashCmdCon.acceptingCmd = false
				m.slashCmdCon.openCmd = ""

				if len(m.chatCentre.prettyHistory()) == 0 {
					m.viewport.SetContent(subtleStyle.Render("press 'ctrl+e' to start chat"))
				} else {
					m.viewport.SetContent(lipgloss.JoinVertical(
						lipgloss.Top,
						strings.Join(m.chatCentre.prettyHistory(), "\n"),
					))
				}

				return m, nil
			}

			if m.chatCentre.fileInpBoxOpen {
				m.chatCentre.fileInpBoxOpen = false
				return m, nil
			}

			CHandler.session.Close()
			return m, tea.Quit

		case "ctrl+e":
			if !m.textarea.Focused() {
				m.textarea.Focus()
			} else {
				m.textarea.Blur()
			}

			return m, cmd

		case "ctrl+x":
			if m.chatCentre.isAgentWorking {
				m.chatCentre.currentAgentRes.agentBgActivity.done(false)
				m.viewport.GotoBottom()
				m.chatCentre.isAgentWorking = false

				return m, doTick50()
			}

		case "alt+enter":
			if m.chatCentre.isAgentWorking {
				return m, cmd
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

		case "pgup", "pgdown", "up", "down":
			if !m.textarea.Focused() {
				m.viewport, vpCmd = m.viewport.Update(msg)
			} else {
				m.textarea, taCmd = m.textarea.Update(msg)
			}

			if m.chatCentre.fileInpBoxOpen {
				m.chatCentre.fileInpBox.vp, fvpCmd = m.chatCentre.fileInpBox.vp.Update(msg)
			}

		case "left", "right":
			if m.slashCmdCon.openCmd == "/help" {
				switch msg.String() {
				case "left":
					helpCmdTabState = 0
					m.viewport.SetContent(helpCmdComponent())
					return m, nil

				case "right":
					helpCmdTabState = 1
					m.viewport.SetContent(helpCmdComponent())
					return m, nil
				}
			}

		case "/":
			if !m.textarea.Focused() && !m.chatCentre.fileInpBoxOpen {
				m.slashCmdCon = newSlashCmdCtrl(0, 0)
				m.slashCmdCon.acceptingCmd = true
				m.slashCmdCon.updateMatches()

				return m, nil
			}

		case "enter":
			if m.chatCentre.fileInpBoxOpen {
				if len(m.chatCentre.fileInpBox.matches) > 0 {
					file := m.chatCentre.fileInpBox.matches[0]
					if m.chatCentre.fileInpBox.prompt == "rm " {
						m.chatCentre.removeFileForContext(file.Path)
						m.chatCentre.fileInpBoxOpen = false

						return m, cmd
					}

					if !file.IsDir {
						m.chatCentre.addFileForContext(file.Path)
						m.chatCentre.fileInpBoxOpen = false
					}
				}

			}

			if m.slashCmdCon.acceptingCmd {
				if m.slashCmdCon.topMatch.name == "" {
					return m, nil
				}

				m.slashCmdCon.acceptingCmd = false

				if m.slashCmdCon.topMatch.name == "/model" {
					m, cmds := CHandler.handleGetModels(m)
					return m, cmds
				}

				if m.slashCmdCon.topMatch.comp != nil {
					m.viewport.SetContent(m.slashCmdCon.topMatch.comp())
				}
				return m, nil
			}

		case "?":
			if !m.textarea.Focused() {
				m.slashCmdCon.openCmd = "/help"
				m.viewport.SetContent(helpCmdComponent())
			}

		case "tab":
			c := m.chatCentre
			if c.fileInpBoxOpen && len(c.fileInpBox.matches) > 0 {
				c.fileInpBox.ti.SetValue(c.fileInpBox.matches[0].Path)
				c.fileInpBox.ti.CursorEnd()
				c.fileInpBox.vp.GotoTop()
			}

			if m.slashCmdCon.acceptingCmd && len(m.slashCmdCon.matches) > 0 {
				if m.slashCmdCon.topMatch.name == "" {
					return m, nil
				}

				if m.slashCmdCon.topMatch.comp != nil {
					m.slashCmdCon.ti.SetValue(m.slashCmdCon.topMatch.name)
					m.slashCmdCon.ti.CursorEnd()
				}
				return m, nil
			}

			m.chatCentre = c

		case "@":
			if !m.textarea.Focused() && !m.slashCmdCon.acceptingCmd {
				m.chatCentre.fileInpBox = newFileInputBox(0, 0)
				m.chatCentre.fileInpBoxOpen = true

				return m, nil
			}
		}
	}

	if m.slashCmdCon.acceptingCmd && !strings.HasPrefix(m.slashCmdCon.ti.Value(), "/") {
		m.slashCmdCon.ti.SetValue("/" + m.slashCmdCon.ti.Value())
	}

	if m.chatCentre.fileInpBoxOpen {
		ti := m.chatCentre.fileInpBox.ti
		vp := m.chatCentre.fileInpBox.vp

		if ti.Value() == "rm " {
			ti = m.chatCentre.fileInpBox.setPrompt(ti, "rm ")
			ti.Reset()
		}

		matches, err := utils.ListMatches(ti.Value())

		if err != nil {
			vp.GotoTop()
			vp.SetContent(err.Error())
		} else {
			listItemStyle := lipgloss.NewStyle().MarginLeft(2)
			var mList []string
			for _, e := range matches {
				if e.IsDir {
					mList = append(mList, listItemStyle.Bold(true).Foreground(accentColor).Render(e.Name))
				} else {
					mList = append(mList, listItemStyle.Render(e.Name))
				}
			}
			vp.SetContent(strings.Join(mList, "\n"))
		}

		m.chatCentre.fileInpBox.ti = ti
		m.chatCentre.fileInpBox.vp = vp
		m.chatCentre.fileInpBox.matches = matches
	}

	if m.chatCentre.sendingPrompt {
		sendPromptState = sendPromptBtnStates[1]
	} else {
		sendPromptState = sendPromptBtnStates[0]
	}

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	if m.chatCentre.fileInpBoxOpen {
		m.chatCentre.fileInpBox.vp, fvpCmd = m.chatCentre.fileInpBox.vp.Update(msg)
		m.chatCentre.fileInpBox.ti, ftiCmd = m.chatCentre.fileInpBox.ti.Update(msg)
	}

	if m.slashCmdCon.acceptingCmd && !m.textarea.Focused() && !m.chatCentre.fileInpBoxOpen {
		m.slashCmdCon.ti, ciCmd = m.slashCmdCon.ti.Update(msg)
		m.slashCmdCon.updateMatches()
	}

	return m, tea.Batch(taCmd, vpCmd, ciCmd, mniCmd, fvpCmd, ftiCmd, cmd)
}


func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing screen..."
	}

	// Vertical height allocations
	headerHeight := 2
	footerHeight := 1
	middleHeight := m.height - headerHeight - footerHeight

	// Assemble rows
	middle := lipgloss.JoinHorizontal(lipgloss.Top, m.chatUi(middleHeight))

	screen := lipgloss.JoinVertical(lipgloss.Left, m.headerBar(headerHeight), middle, m.footerBar(footerHeight))

	return zone.Scan(screen)
}

func (m *Model) handleErrMsg(msg errMsg) (tea.Model, tea.Cmd) {
	m.chatCentre.isAgentWorking = false
	m.chatCentre.sendingPrompt = false
	sendPromptState = sendPromptBtnStates[0]
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

	m.textarea = createTextArea(msg.Width, taHeight, m.textarea.Value())

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

		switch {
		case zone.Get("text-area").InBounds(msg):
			m.textarea.Focus()
			return m, nil

		case zone.Get("add-file-btn").InBounds(msg):
			m.slashCmdCon.acceptingCmd = false

			m.chatCentre.fileInpBox = newFileInputBox(0, 0)
			m.chatCentre.fileInpBoxOpen = true
			return m, nil

		case zone.Get("model-selector-btn").InBounds(msg):
			m.slashCmdCon.acceptingCmd = false
			m.chatCentre.fileInpBoxOpen = false
			m.slashCmdCon.openCmd = "/model"
			return m, nil

		case zone.Get("mode-toggle-btn").InBounds(msg):
			config.Config.Mode.Toggle()
			return m, nil

		case zone.Get("send-prompt-btn").InBounds(msg):
			return m, sendPromptKeyMsg()

		case zone.Get("retry-button").InBounds(msg):
			m, cmds := CHandler.handlePrompt(m)

			return m, tea.Batch(cmds, doTick20())

		case zone.Get("helpCmd-general-btn").InBounds(msg):
			helpCmdTabState = 0
			m.viewport.SetContent(helpCmdComponent())
			return m, nil

		case zone.Get("helpCmd-commands-btn").InBounds(msg):
			helpCmdTabState = 1
			m.viewport.SetContent(helpCmdComponent())
			return m, nil

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
		m.viewport.SetContent(strings.Join(m.chatCentre.prettyHistory(), "\n"))
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
		sendPromptState = sendPromptBtnStates[0]
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
