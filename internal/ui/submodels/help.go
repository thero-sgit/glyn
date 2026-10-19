package submodels

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

var (
	helpCmdGeneralTabContent = lipgloss.NewStyle().
					Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				library.BoldStyle.Render("esc")+"/"+library.BoldStyle.Render("ctrl+c")+" to quit glyn",
				library.BoldStyle.Render("/")+" display '/' command input screen",

				library.SubtleStyle.MarginTop(1).Render("AI interaction & config"),
				library.BoldStyle.Render("alt+enter")+"/click "+library.SendPromptBtnStates[0]+" send prompt",
				library.BoldStyle.Render("@")+"/click "+library.BoldStyle.Render("@")+" to add files",
				"click "+library.LabelValueBand("\u27E1", "`model-id`", "")+" to view models list",
				"click "+library.LabelValueBand("?/"+"\u2726", "`mode`", "")+" to switch mode",

				library.SubtleStyle.MarginTop(1).Render("In chat screen"),
				library.BoldStyle.Render("pgup, pgdn, up, down arrows")+"/"+library.BoldStyle.Render("mouse")+" to scroll",
				library.BoldStyle.Render("ctrl+e")+" to focus prompt text input",
				library.BoldStyle.Render("ctrl+x")+"/click "+library.SendPromptBtnStates[1]+" to interupt session",
			),
		)

	helpCmdCommandsTabContent = lipgloss.NewStyle().
					Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				library.BoldStyle.Render("/help")+" display helpful keyboard shortcuts",
				library.BoldStyle.Render("/status")+" display usage and general status",
				library.BoldStyle.Render("/model")+" display models list",
			),
		)
)

type HelpModel struct {
	state    int
	viewport viewport.Model
}

func (h *HelpModel) Init() tea.Cmd {
	return nil
}

func NewHelpModel() tea.Model {
	return &HelpModel{
		viewport: viewport.New(0, 0),
		state: 0,
	}
}

func (h *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.viewport.Width, h.viewport.Height = msg.Width, msg.Height
		return h, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			case zone.Get("helpCmd-general-btn").InBounds(msg):
				h.state = 0
				return h, cmd

			case zone.Get("helpCmd-commands-btn").InBounds(msg):
				h.state = 1
				return h, cmd
			}
		}
			
		h.viewport, cmd = h.viewport.Update(msg)
		return h, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			h.state = 0

		case "right":
			h.state = 1
		}
	}

	return h, nil
}

func (h *HelpModel) View() string {
	var generalTabBtn string
	var commandsTabBtn string
	var content string

	if h.state == 0 {
		generalTabBtn = library.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("General")
		commandsTabBtn = library.SubtleStyle.Render(" Commands ")
		content = helpCmdGeneralTabContent
	} else {
		generalTabBtn = library.SubtleStyle.Render(" General ")
		commandsTabBtn = library.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("Commands")
		content = helpCmdCommandsTabContent
	}

	generalTabBtn = zone.Mark("helpCmd-general-btn", generalTabBtn)
	commandsTabBtn = zone.Mark("helpCmd-commands-btn", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	headerBar := lipgloss.NewStyle().
		MarginBottom(1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				library.BwLabelStyle.Render(" /help "),
				" ",
				tabs,
				" ",
				"Useful Shortcuts & '/' Commands",
				" ",
				library.SubtleStyle.MarginLeft(5).Render("press 'esc' to close   left/right arrows to switch tabs"),
			),
	)

	h.viewport.SetContent(
		lipgloss.JoinVertical(
			lipgloss.Top,
			headerBar,
			content,
		),
	)

	return h.viewport.View()
}
