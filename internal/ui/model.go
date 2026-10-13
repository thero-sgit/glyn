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
	"github.com/thero-sgit/xyn/internal/common"
	"github.com/thero-sgit/xyn/internal/ui/submodels"
)

type TickMsg50 time.Time

func doTick50() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg50(t)
	})
}

type TickMsg20 time.Time

func doTick20() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg20(t)
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

	// Sub Models
	slashCmdCon    submodels.Selector
	helpScreen     submodels.HelpModel
	sessionsScreen submodels.Sessions
	llmsScreen     submodels.LLModels
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
		chatCentre:     newChatCentre(),
		slashCmdCon:    submodels.NewSelector("Command", "/", "/", "<command>", common.ListCmdMatches),
		helpScreen:     submodels.NewHelpModel(),
		sessionsScreen: submodels.NewSessions(),
		llmsScreen:     submodels.NewModelsModel(),
		viewport:       vp,
	}

	model.viewport.SetContent(strings.Join(model.chatCentre.prettyHistory(), "\n"))
	InitHandler()

	return &model
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.MouseMsg:
		return m.handleMouseAction(msg)
	case errMsg:
		return m.handleErrMsg()
	case TickMsg50:
		return m.handleTick50()
	case TickMsg20:
		return m.handleTick20()
	case response:
		return m.handleResponse(msg)
	case sessionInfo:
		return m.handleSessionInfo(msg)
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "IniTializing screen..."
	}

	headerHeight := 2
	footerHeight := 1
	middleHeight := m.height - headerHeight - footerHeight

	middle := lipgloss.JoinHorizontal(lipgloss.Top, m.chatUi(middleHeight))
	screen := lipgloss.JoinVertical(lipgloss.Left, m.headerBar(headerHeight), middle, m.footerBar(footerHeight))
	return zone.Scan(screen)
}

func (m *Model) renderChat() {
	m.viewport.SetContent(strings.Join(m.chatCentre.prettyHistory(), "\n"))
}
