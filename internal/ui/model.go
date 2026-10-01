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
	"github.com/thero-sgit/xyn/internal/ai"
	"github.com/thero-sgit/xyn/internal/config"
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
		return tea.KeyMsg {
			Type: tea.KeyEnter,
			Alt: true,
		}
	}
}

type Model struct {
	width  		   int
	height 		   int
	textarea 	   textarea.Model
	viewport       viewport.Model

	//
	chatCentre     chatCentre

	//
	slashCmdCon    slashCmdCtrl
}

func InitialModel() Model {
	vp := viewport.New(0, 0)

	vp.KeyMap = viewport.KeyMap{
		Up:   key.NewBinding(key.WithKeys("up")),                                                                                                                                                                       
		Down: key.NewBinding(key.WithKeys("down")),                                                                                                                                                                         
		PageUp:  key.NewBinding(key.WithKeys("pgup")),                                                                                                                                                                                 
		PageDown: key.NewBinding(key.WithKeys("pgdn")),                                                                                                                                                                                                
		HalfPageUp:    key.NewBinding(),                                                                                                                                                                                                  
		HalfPageDown:  key.NewBinding(),
	}


	model := Model{
		chatCentre: newChatCentre(),
		viewport: vp,
	}

	model.viewport.SetContent(strings.Join(model.chatCentre.prettyHistory(), "\n"))

	InitHandler()

	return model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd   tea.Cmd
		vpCmd tea.Cmd
		taCmd tea.Cmd
		ciCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		taHeight := 2
		if m.textarea.Length() > 0 {
			taHeight = 6
		}

		m.textarea = createTextArea(msg.Width, taHeight, m.textarea.Value())

        headerHeight := 2
        footerHeight := 1
        middleHeight := m.height - headerHeight - footerHeight
        promptBoxHeight := taHeight + 1

		m.viewport.Height = max(middleHeight - promptBoxHeight, 1)
		m.viewport.Width  = m.width - 4

		config.Config.SetRenderer(m.viewport.Width-2)

		m.chatCentre.updatedWidths(m.viewport.Width)

	case tea.MouseMsg:
        if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m.textarea.Blur()
			switch{
			case zone.Get("text-area").InBounds(msg):
				m.textarea.Focus()
                return m, nil

			case zone.Get("add-context-btn").InBounds(msg):
				m.slashCmdCon.info.acceptingCmd = false
				m.slashCmdCon.info.openCmd = ""
				m.viewport.SetContent("ADD CONTEXT!!")
                return m, nil

			case zone.Get("model-selector-btn").InBounds(msg):
				m.slashCmdCon.info.acceptingCmd = false
				m.slashCmdCon.info.openCmd = ""
				m.viewport.SetContent("MODELL!!")
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

	case errMsg:
		if m.slashCmdCon.info.openCmd == "/models" {
			m.viewport.SetContent(modelsCmdComponent(msg.err.Error(), []ai.GroqModelInfo{}, 0))

			return m, nil
		}

        m.chatCentre.isAgentWorking = false
        m.chatCentre.sendingPrompt  = false
        sendPromptState = sendPromptBtnStates[0]
        m.chatCentre.currentUserPrompt.err()
        m.viewport.SetContent(strings.Join(append(m.chatCentre.prettyHistory(), m.chatCentre.currentUserPrompt.getPretty()), "\n"))

        return m, nil

    case tickMsg50:
        if m.chatCentre.isAgentWorking {
            m.chatCentre.currentAgentRes.agentBgActivity.animate()
            cmd = doTick50()
        } else {
            m.viewport.SetContent(strings.Join(m.chatCentre.prettyHistory(), "\n"))
        }

        return m, cmd

    case tickMsg20:
        if m.chatCentre.sendingPrompt {
			m.viewport.SetContent(strings.Join(append(m.chatCentre.prettyHistory(), m.chatCentre.currentUserPrompt.getPretty()), "\n"))
            m.chatCentre.currentUserPrompt.animate()
            cmd = doTick20()
        }

        return m, cmd

    case response:
        var c tea.Cmd
        if m.chatCentre.sendingPrompt {
            m.chatCentre.sendingPrompt = false
            m.chatCentre.isAgentWorking = true
            m.chatCentre.currentUserPrompt.sent()
            m.chatCentre.history = append(m.chatCentre.history, m.chatCentre.currentUserPrompt)
            c = doTick50()
        }

        if msg.chunk.EOS {
            m.chatCentre.isAgentWorking = false
            m.chatCentre.currentAgentRes.agentBgActivity.done(false)
            sendPromptState = sendPromptBtnStates[0]
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
        m.viewport.GotoBottom()

        return m, tea.Batch(awaitResponse(CHandler.session.Ctx, msg.c), c)

    case sessionInfo:
        m.chatCentre.sessionName += msg.chunk.Data

        if msg.chunk.EOS {
            return m, nil
        }

        return m, awaitSessionInfo(CHandler.session.Ctx, msg.c)


	case models:
		models := msg.m
		m.viewport.SetContent(modelsCmdComponent("", models, m.viewport.Width))

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c" :
			CHandler.session.Close()
			return m, tea.Quit

		case "esc":
			if m.slashCmdCon.info.acceptingCmd || m.slashCmdCon.info.openCmd != "" {
				m.slashCmdCon.info.acceptingCmd = false
				m.slashCmdCon.info.openCmd = ""

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

			CHandler.session.Close()
			return m, tea.Quit

		case "ctrl+e":
			if !m.textarea.Focused() {
				m.textarea.Focus()
				m.slashCmdCon.info.acceptingCmd = false
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
				m.chatCentre.currentUserPrompt       = newUserPrompt(prompt, m.viewport.Width)
				m.chatCentre.currentAgentRes   		 = newAgentResponse(m.viewport.Width)
				m.chatCentre.currentAgentRes.index   = max(0, len(m.chatCentre.history)-1)
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

			if m.slashCmdCon.info.acceptingCmd {
				if msg.String() == "up" {
					if m.slashCmdCon.info.highlightedIndex == 0 {
						m.slashCmdCon.info.highlightedIndex = m.slashCmdCon.info.optsLen - 1
					} else {
						m.slashCmdCon.info.highlightedIndex = m.slashCmdCon.info.highlightedIndex - 1
					}
				}

				if msg.String() == "down" {
					m.slashCmdCon.info.highlightedIndex = (m.slashCmdCon.info.highlightedIndex + 1) % m.slashCmdCon.info.optsLen
				}

				
				if len(m.slashCmdCon.info.ta.Value()) <= 1 {
					 m.slashCmdCon.info.ta.Reset()
					 m.slashCmdCon.info.ta.Placeholder = m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].getName()
				}
				m.viewport.SetContent(lipgloss.JoinVertical(
					lipgloss.Top,
					viewSelector(&m.slashCmdCon),
				))
			}

			return m, nil

		case "left", "right":
			if m.slashCmdCon.info.openCmd == "/help" {
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
			if !m.textarea.Focused() {
				m.slashCmdCon.info.acceptingCmd = true
				m.slashCmdCon              = newSlashCmdCtrl()			
			}

		case "enter":
			if m.slashCmdCon.info.acceptingCmd && m.slashCmdCon.info.optsLen > 0 && !m.textarea.Focused() {
				m.slashCmdCon.info.acceptingCmd = false
				m.slashCmdCon.info.openCmd = m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].getName()

				if m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].getName() == "/models" {
					m.viewport.SetContent(modelsCmdComponent("Loading...", []ai.GroqModelInfo{}))
					m, cmds := CHandler.handleGetModels(m)

					return m, cmds
				}

				m.viewport.SetContent(m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].component())
			}

		case "?":
			if !m.textarea.Focused() {
				m.slashCmdCon.info.openCmd = "/help"
				m.viewport.SetContent(helpCmdComponent())
			}
		}		
	}

	if m.slashCmdCon.info.acceptingCmd {
		m.slashCmdCon.info.options = slashCommands
		m.viewport.SetContent(lipgloss.JoinVertical(
			lipgloss.Top,
			viewSelector(&m.slashCmdCon),
		))
	}

	if m.chatCentre.sendingPrompt {
		sendPromptState = sendPromptBtnStates[1]
	} else {
		sendPromptState = sendPromptBtnStates[0]
	}

	m.textarea, taCmd       = m.textarea.Update(msg)
	m.viewport, vpCmd       = m.viewport.Update(msg)
	m.slashCmdCon.info.ta, ciCmd = m.slashCmdCon.info.ta.Update(msg)
	return m, tea.Batch(taCmd, vpCmd, ciCmd, cmd)
}

func (m Model) View() string {
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