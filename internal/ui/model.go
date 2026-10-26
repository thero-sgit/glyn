package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ai"
	"github.com/thero-sgit/glyn/internal/common"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/submodels"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

type Model struct {
	width          int
	height         int
	mHeight        int
	inputBoxHeight int
	textarea       textarea.Model

	//
	handler api.Handler

	//
	slashCmdCon    components.Selector
	fileSelector   components.FileSelector

	// Screens
	chatScreen     tea.Model
	helpScreen     tea.Model
	sessionsScreen tea.Model
	llmsScreen     tea.Model

	focused        *tea.Model
}

func InitialModel() *Model {
	s := ai.NewSession()
	h := api.NewHandler(s)

	model := Model{
		handler:        h,
		slashCmdCon:    components.NewSelector("Command", "/", "/", "<command>", common.ListCmdMatches),
		fileSelector:   components.NewFileSelector(),
		helpScreen:     submodels.NewHelpModel(),
		sessionsScreen: submodels.NewSessionsScreen(),
		llmsScreen:     submodels.NewModelsModel(),
	}

	model.chatScreen = submodels.NewChatScreen(&model.handler)
	model.focused = &model.chatScreen

	return &model
}

func (m *Model) Init() tea.Cmd {
	return m.handler.HandleGetModels()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.MouseMsg:
		return m.handleMouseAction(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case api.ErrorMsg:
		return m.handleErrMsg(msg)

	case api.ModelsCmd:
		return m.handleModelsCmd(msg)

	case api.TickMsg20:
		return m.handleTick20(msg)

	case api.TickMsg50:
		return m.handleTick50(msg)

	case api.ResponseChunkCmd:
		return m.handleResponseChunkCmd(msg)

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

	//
	focused := *m.focused
	focused.Update(tea.WindowSizeMsg{
		Width:  m.width,
		Height: m.mHeight,
	})

	content := lipgloss.NewStyle().Padding(0, 2, 0, 2).Height(m.mHeight).Render(focused.View())
	screen := lipgloss.JoinVertical(lipgloss.Left, m.headerBar(headerHeight), content, inputBox, m.footerBar(footerHeight))

	return zone.Scan(screen)
}
