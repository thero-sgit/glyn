package submodels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

var (
	helpCmdGeneralTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				stylesandcomponents.BoldStyle.Render("esc") + "/" + stylesandcomponents.BoldStyle.Render("ctrl+c") + " to quit xyn",
				stylesandcomponents.BoldStyle.Render("/") + " display '/' command input screen",

				stylesandcomponents.SubtleStyle.MarginTop(1).Render("AI interaction & config"),
				stylesandcomponents.BoldStyle.Render("alt+enter") + "/click " + stylesandcomponents.SendPromptBtnStates[0] + " send prompt",
				stylesandcomponents.BoldStyle.Render("@") + "/click " + stylesandcomponents.BoldStyle.Render("@") + " to add files",
				"click " + stylesandcomponents.LabelValueBand("\u27E1", "`model-id`") + " to view models list",
				"click " + stylesandcomponents.LabelValueBand("?/" + "\u2726", "`mode`") + " to switch mode",
				
				stylesandcomponents.SubtleStyle.MarginTop(1).Render("In chat screen"),
				stylesandcomponents.BoldStyle.Render("pgup, pgdn, up, down arrows") + "/" + stylesandcomponents.BoldStyle.Render("mouse") + " to scroll",
				stylesandcomponents.BoldStyle.Render("ctrl+e") + " to focus prompt text input",
				stylesandcomponents.BoldStyle.Render("ctrl+x") + "/click " + stylesandcomponents.SendPromptBtnStates[1] + " to interupt session",
			),
		)

	helpCmdCommandsTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				stylesandcomponents.BoldStyle.Render("/help") + " display helpful keyboard shortcuts",
				stylesandcomponents.BoldStyle.Render("/status") + " display usage and general status",
				stylesandcomponents.BoldStyle.Render("/model") + " display models list",
			),
		)
)

type HelpModel struct {
	state   int
	Focused bool
}


func (h *HelpModel) Init() tea.Cmd {
	return nil
}

func NewHelpModel() HelpModel {
	return HelpModel{ state: 0 }
}

func (h *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			case zone.Get("helpCmd-general-btn").InBounds(msg):
				h.state = 0
			
			case zone.Get("helpCmd-commands-btn").InBounds(msg):
				h.state = 1
			}
		}

		return h, nil

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
	var generalTabBtn  string
	var commandsTabBtn string
	var content        string

	if h.state == 0 {
		generalTabBtn  = stylesandcomponents.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("General")
		commandsTabBtn = stylesandcomponents.SubtleStyle.Render(" Commands ")
		content 	   = helpCmdGeneralTabContent
	} else {
		generalTabBtn  = stylesandcomponents.SubtleStyle.Render(" General ")
		commandsTabBtn = stylesandcomponents.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("Commands")
		content        = helpCmdCommandsTabContent
	}

	generalTabBtn  = zone.Mark("helpCmd-general-btn", generalTabBtn)
	commandsTabBtn = zone.Mark("helpCmd-commands-btn", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	headerBar := lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinHorizontal(
			lipgloss.Left, 
			stylesandcomponents.BwLabelStyle.Render(" /help "),
			" ",
			tabs,
			" ",
			"Useful Shortcuts & '/' Commands",
			" ",
			stylesandcomponents.SubtleStyle.MarginLeft(5).Render("press 'esc' to close   left/right arrows to switch tabs"),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		headerBar,
		content,
	)
}
