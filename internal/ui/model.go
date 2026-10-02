package ui

import (
	"errors"
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
	modelsSelCon   modelsSelectorCtrl
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

			case zone.Get("add-file-btn").InBounds(msg):
				m.slashCmdCon.info.acceptingInp  = false
				m.modelsSelCon.info.acceptingInp = false
				
				m.chatCentre.fileInpBox 	= newFileInputBox(0, 0)
				m.chatCentre.fileInpBoxOpen = true
                return m, nil

			case zone.Get("model-selector-btn").InBounds(msg):
				m.slashCmdCon.info.acceptingInp = false
				m.modelsSelCon.info.acceptingInp = true
				m.slashCmdCon.info.openCmd = "/model"
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
		if m.slashCmdCon.info.openCmd == "/model" {
			_, content := modelsCmdComponent(m.modelsSelCon, msg.err.Error(), []ai.GroqModelInfo{})
			m.viewport.SetContent(content)

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
		info, content := modelsCmdComponent(m.modelsSelCon, "", models)
		m.modelsSelCon.info = info
		m.modelsSelCon.info.acceptingInp = true
		m.modelsSelCon.info.ti.Focus()

		m.viewport.SetContent(content)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c" :
			CHandler.session.Close()
			return m, tea.Quit

		case "esc":
			if m.slashCmdCon.info.acceptingInp || m.slashCmdCon.info.openCmd != "" {
				m.slashCmdCon.info.acceptingInp  = false
				m.modelsSelCon.info.acceptingInp = false
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

			m.chatCentre.fileInpBox.vp, fvpCmd = m.chatCentre.fileInpBox.vp.Update(msg)

			info, err := func() (selectorInfo, error) {
				if m.slashCmdCon.info.acceptingInp {
					return m.slashCmdCon.info, nil
				}

				if m.modelsSelCon.info.acceptingInp {
					return m.modelsSelCon.info, nil
				}

				return selectorInfo{}, errors.New("")
			}()

			if err == nil && info.optsLen > 1 {
				if msg.String() == "up" {
					if info.highlightedIndex == 0 {
						info.highlightedIndex = info.optsLen - 1
					} else {
						info.highlightedIndex = info.highlightedIndex - 1
					}
				}

				if msg.String() == "down" {
					info.highlightedIndex = (info.highlightedIndex + 1) % max(info.optsLen, 1)
				}

				
				if len(info.ti.Value()) <= 1 {
					 info.ti.Reset()
					 info.ti.Placeholder = info.options[info.highlightedIndex].getInfo().name
				}


				if m.slashCmdCon.info.acceptingInp {
					m.slashCmdCon.info = info
				} else {
					m.modelsSelCon.info = info
				}
			}

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
			if !m.textarea.Focused() && !m.chatCentre.fileInpBoxOpen {
				m.slashCmdCon.info.acceptingInp = true
				m.modelsSelCon.info.acceptingInp = false
				m.slashCmdCon              		= newSlashCmdCtrl()			
			}

		case "enter":
			if m.chatCentre.fileInpBoxOpen {
				m.chatCentre.fileInpBoxOpen = false
			}

			if m.slashCmdCon.info.acceptingInp && m.slashCmdCon.info.optsLen > 0 && !m.textarea.Focused() {
				m.slashCmdCon.info.acceptingInp = false
				m.slashCmdCon.info.openCmd = m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].getInfo().name

				if m.slashCmdCon.info.openCmd == "/model" {
					m.modelsSelCon = newModelsSelectorCtrl()
					m, cmds := CHandler.handleGetModels(m)

					return m, cmds
				}

				m.viewport.SetContent(m.slashCmdCon.info.options[m.slashCmdCon.info.highlightedIndex].getInfo().comp())
			}

		case "?":
			if !m.textarea.Focused() {
				m.slashCmdCon.info.openCmd = "/help"
				m.viewport.SetContent(helpCmdComponent())
			}

		case "tab":
			c := m.chatCentre
			if c.fileInpBoxOpen && len(c.fileInpBox.matches) > 0 {
				c.fileInpBox.ti.SetValue(c.fileInpBox.matches[0].Path)
				c.fileInpBox.ti.CursorEnd()
				c.fileInpBox.vp.GotoTop()
			}

			m.chatCentre = c

		case "@":
			if !m.textarea.Focused() {
				m.chatCentre.fileInpBox 	= newFileInputBox(0, 0)
				m.chatCentre.fileInpBoxOpen = true

				return m, nil
			}
		}
	}

	if (m.slashCmdCon.info.acceptingInp || m.modelsSelCon.info.acceptingInp) {
		if m.slashCmdCon.info.acceptingInp {
			m.viewport.SetContent(viewSelector(&m.slashCmdCon))
		} else {
			_, content := modelsCmdComponent(m.modelsSelCon, "", []ai.GroqModelInfo{})
			m.viewport.SetContent(content)
		}
	}

	if m.chatCentre.fileInpBoxOpen {
		ti := m.chatCentre.fileInpBox.ti
		vp := m.chatCentre.fileInpBox.vp

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

	m.textarea, taCmd       	   = m.textarea.Update(msg)
	m.viewport, vpCmd       	   = m.viewport.Update(msg)

	if m.chatCentre.fileInpBoxOpen {
		m.chatCentre.fileInpBox.vp, fvpCmd = m.chatCentre.fileInpBox.vp.Update(msg)
		m.chatCentre.fileInpBox.ti, ftiCmd = m.chatCentre.fileInpBox.ti.Update(msg)
	}	
	
	if !m.textarea.Focused() {
		m.slashCmdCon.info.ti, ciCmd   = m.slashCmdCon.info.ti.Update(msg)
		m.modelsSelCon.info.ti, mniCmd = m.modelsSelCon.info.ti.Update(msg)
	}

	return m, tea.Batch(taCmd, vpCmd, ciCmd, mniCmd, fvpCmd, ftiCmd, cmd)
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