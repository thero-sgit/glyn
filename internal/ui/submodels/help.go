package submodels

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

var (
	helpCmdGeneralTabContent = lipgloss.JoinVertical(
		lipgloss.Top,
		library.SubtleStyle.Render("Chat"),
		library.BoldStyle.Render("enter")+" send prompt    "+
			library.BoldStyle.Render("alt+enter / ctrl+j")+" insert a newline",
		library.BoldStyle.Render("ctrl+e")+" focus or unfocus the prompt",
		library.BoldStyle.Render("Mouse wheel")+" scroll chat history",

		library.SubtleStyle.MarginTop(1).Render("Files & settings"),
		library.BoldStyle.Render("@")+" open the file picker; type to filter, arrows to choose, "+
			"enter/tab to select (enter folders to browse)",
		"Click "+library.LabelValueBand("\u27E1", "`model`", "")+" to choose a model and effort",
		"Click "+library.LabelValueBand("? / \u2726", "`mode`", "")+" to switch ask/agent mode",

		library.SubtleStyle.MarginTop(1).Render("Navigation"),
		library.BoldStyle.Render("/")+" browse commands; "+library.BoldStyle.Render("enter")+" run selection, "+
			library.BoldStyle.Render("tab")+" complete",
		library.BoldStyle.Render("?")+" open help    "+
			library.BoldStyle.Render("esc")+" close screen / return to chat    "+
			library.BoldStyle.Render("ctrl+c")+" quit",
	)

	helpCmdCommandsTabContent = lipgloss.JoinVertical(
		lipgloss.Top,
		library.BoldStyle.Render("/help")+" open keyboard shortcuts and command help",
		library.BoldStyle.Render("/sessions")+" browse and resume a previous chat",
		library.BoldStyle.Render("/model")+" browse available models and adjust model effort",
		library.SubtleStyle.MarginTop(1).Render(
			"Type / to open the command picker, then type to filter. "+
				"Use up/down to select, tab to complete, and enter to run.",
		),
		library.SubtleStyle.Render(
			"In the sessions list, use up/down to choose a chat and enter to resume it.",
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
		state:    0,
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
				"Shortcuts & Commands",
				" ",
				library.SubtleStyle.MarginLeft(5).Render("esc close   left/right switch tabs"),
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
