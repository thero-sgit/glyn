package ui

import (
	"time"
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
	mHeight  int
	textarea textarea.Model
	viewport viewport.Model

	//
	chatCentre chatCentre

	// Sub Models
	slashCmdCon    submodels.Selector
	helpScreen     tea.Model
	sessionsScreen tea.Model
	llmsScreen     tea.Model

	focused        *tea.Model
}

func InitialModel() *Model {
	model := Model{
		chatCentre:     newChatCentre(),
		slashCmdCon:    submodels.NewSelector("Command", "/", "/", "<command>", common.ListCmdMatches),
		helpScreen:     submodels.NewHelpModel(),
		sessionsScreen: submodels.NewSessions(),
		llmsScreen:     submodels.NewModelsModel(),
	}

	model.focused = &model.helpScreen
	InitHandler()

	return &model
}

func (m *Model) Init() tea.Cmd {
	var cmd tea.Cmd
	m, cmd = CHandler.handleGetModels(m)
	return cmd
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	default:
		focused := *m.focused
		focused.Update(msg)
		return m, nil
	}
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "IniTializing screen..."
	}

	//
	headerHeight := 2
	footerHeight := 1
	middleHeight := m.height - headerHeight - footerHeight

	// 
	inputBox := m.inputBox(middleHeight)
	focused := *m.focused

	content := lipgloss.NewStyle().Padding(0, 2, 0, 2).Height(m.mHeight).Render(focused.View())
	screen := lipgloss.JoinVertical(lipgloss.Left, m.headerBar(headerHeight), content, inputBox, m.footerBar(footerHeight))

	return zone.Scan(screen)
}
