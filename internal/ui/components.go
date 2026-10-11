package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ai"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/utils"
)

const xynLargeText = 
`██╗ ██╗██╗ ██╗██████╗
 ████╔╝██████║██╔═██║
██╔═██╗╚═══██║██║ ██║
╚═╝ ╚═╝    ╚═╝╚═╝ ╚═╝`

const codeLargeText = 
`██████╗██████╗    ██╗██████╗█████╗
██╔═██║██╔═██║██████║██╔═██║██╔═██╗
██████║██║ ██║██╔═██║█████╔╝██║ ╚═╝
    ██║██████║██████║██████╗███╗
    ╚═╝╚═════╝╚═════╝╚═════╝╚══╝`


var (
	
	helpCmdTabState   = 0
	helpCmdGeneralTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				utils.BoldStyle.Render("esc") + "/" + utils.BoldStyle.Render("ctrl+c") + " to quit xyn",
				utils.BoldStyle.Render("/") + " display '/' command input screen",

				utils.SubtleStyle.MarginTop(1).Render("AI interaction & config"),
				utils.BoldStyle.Render("alt+enter") + "/click " + sendPromptBtnStates[0] + " send prompt",
				utils.BoldStyle.Render("@") + "/click " + utils.BoldStyle.Render("@") + " to add files",
				"click " + labelValueBand("\u27E1", "`model-id`") + " to view models list",
				"click " + labelValueBand("?/" + "\u2726", "`mode`") + " to switch mode",
				
				utils.SubtleStyle.MarginTop(1).Render("In chat screen"),
				utils.BoldStyle.Render("pgup, pgdn, up, down arrows") + "/" + utils.BoldStyle.Render("mouse") + " to scroll",
				utils.BoldStyle.Render("ctrl+e") + " to focus prompt text input",
				utils.BoldStyle.Render("ctrl+x") + "/click " + sendPromptBtnStates[1] + " to interupt session",
			),
		)

	helpCmdCommandsTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				utils.BoldStyle.Render("/help") + " display helpful keyboard shortcuts",
				utils.BoldStyle.Render("/status") + " display usage and general status",
				utils.BoldStyle.Render("/model") + " display models list",
			),
		)

	slashCommands = []slashCommand {
		{ name: "/help", pretty: "/help  ", comp: helpCmdComponent},
		{ name: "/status", pretty: "/status", comp: statusCmdComponent},
		{ name: "/model", pretty: "/model"},
	}
	slashCmdNames = func() []string {
		var names []string
		for _, slashCmd := range slashCommands { names = append(names, slashCmd.name)}
		return names
	}()

	sendPromptBtnStates = []string {
		utils.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("send " + "\u27A4"),
		utils.BwLabelStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("stop " + "\u25A0"),
	}

	sendPromptState = sendPromptBtnStates[0]
)

func bunner() string {
	inf := lipgloss.NewStyle().MarginLeft(2).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			utils.BoldStyle.Render("xyn ") + utils.SubtleStyle.Bold(true).Render(config.Config.Version),
			utils.BoldStyle.Foreground(utils.AccentColor).Render("\u27E1 ") + utils.SubtleStyle.Render(config.Config.Model),
			utils.BoldStyle.Foreground(utils.AccentColor).Render("@ ") + utils.SubtleStyle.Underline(true).Render("console.groq.com"),
			utils.BoldStyle.Foreground(utils.AccentColor).Render("~") + utils.SubtleStyle.Render(config.Config.WorkingDir),
		),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(utils.AccentColor).Render(xynLargeText), utils.SubtleStyle.Render(codeLargeText)),
		inf,
	)
}

type slashCommand struct {
	name   string
	pretty string
	comp   func() string
}

func helpCmdComponent() string {
	var generalTabBtn  string
	var commandsTabBtn string
	var content        string
	if helpCmdTabState == 0 {
		generalTabBtn  = utils.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("General")
		commandsTabBtn = utils.SubtleStyle.Render(" Commands ")
		content 	   = helpCmdGeneralTabContent
	} else {
		generalTabBtn  = utils.SubtleStyle.Render(" General ")
		commandsTabBtn = utils.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("Commands")
		content        = helpCmdCommandsTabContent
	}

	generalTabBtn  = zone.Mark("helpCmd-general-btn", generalTabBtn)
	commandsTabBtn = zone.Mark("helpCmd-commands-btn", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	headerBar := lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinHorizontal(
			lipgloss.Left, 
			utils.BwLabelStyle.Render(" /help "),
			" ",
			tabs,
			" ",
			"Useful Shortcuts & '/' Commands",
			" ",
			utils.SubtleStyle.MarginLeft(5).Render("press 'esc' to close   left/right arrows to switch tabs"),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		headerBar,
		content,
	)
}

func statusCmdComponent() string {
	cmdLabel := utils.BwLabelStyle.Render(" /status ")

	tokenUsage := "Token usage (34,500 / 200,000 tokens — 17.2%)"
	header 	   := lipgloss.JoinHorizontal(lipgloss.Left, cmdLabel, " ", tokenUsage)

	tokenUsageProgressBar := tokenUsageProgessBar(5)		

	tokenUsageBreakdown := fmt.Sprint(
		"└ Breakdown by Category:\n",
		"	  Conversation History ......... 18,200 tokens (52.8%)\n",
		"	  Loaded Files (3) ............. 9,400 tokens (27.2%)\n",
		"	  System & Instructions ........ 4,100 tokens (11.9%)\n",
		"	  Tools Use & Skills ........... 2,800 tokens ( 8.1%)\n",
	)

	context := fmt.Sprint(
		"└ Context:\n",
		"	 • Model (openai/gpt-oss-120b) Token Limit Resets in 14h 2m 33s\n",
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		tokenUsageProgressBar,
		tokenUsageBreakdown,
		context,
	)
}

func modelCard(modelInfo ai.GroqModelInfo) string {
	header := lipgloss.NewStyle().Render(
			utils.BoldStyle.Render(utils.SubtleStyle.Render("\u27E1")) +
			" " + modelInfo.Name + 
			" (" + modelInfo.Id + ") " + 
			utils.SubtleStyle.Render(modelInfo.OwnedBy),
	)

	c := map[string]string {
		"tools": 	 "\u2692",
		"reasoning": "\u223f",
	}

	var labels []string
	for _, s := range modelInfo.SupportedFeatures {
		val, ok := c[s]
		if ok {
			labels = append(
				labels,
				utils.SubtleStyle.Render(val + " ") + s + "  ",
			)
		}
	}

	capabilities := lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			utils.SubtleStyle.Render("Capabilities"),
			lipgloss.JoinHorizontal(lipgloss.Left, labels...),
		),
	)

	card := lipgloss.NewStyle().MarginBottom(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			capabilities,
		),
	)

	return card
}

func createTextArea(width, height int, value string,) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = ` Try "how does <filename> work?"`
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	
	ta.SetWidth(width - 20)
	ta.SetValue(value)
	ta.SetHeight(height)

	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE"))

	return ta
}

func dirAndSessionLabel(directoryPath, sessionName string) string {
	directoryPath = lipgloss.NewStyle().
		Render(lipgloss.NewStyle().Foreground(utils.AccentColor).Render("~") + utils.SubtleStyle.Faint(true).Render(directoryPath))
	
	sColor := func() string {
		if sessionName == "" {
			return "#b1b2b1"
		}

		return "#06753a"
	}()

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(sColor)).
		Render(" • ")

	sessionName = lipgloss.NewStyle().
		PaddingRight(2).
		Render(sessionName)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, directoryPath, separator, sessionName)

	return lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		PaddingRight(2).
		Render(joined)
}

func labelValueBand(label, value string) string {
	label = utils.HighlighedStyle.
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Render(label )

	value = utils.HighlighedStyle.
		Bold(true).
		PaddingRight(1).
		Render(value)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, label, value)

	output := utils.BwLabelStyle.
		AlignHorizontal(lipgloss.Left).
		Render(joined)

	return output
}

func tokenUsageProgessBar(usage float32) string {
	widthAdj  := 25
	usageBars := int(usage * (float32(widthAdj)/100))

	var filled string 

	for i := 0; i < usageBars; i++ {
		filled += "■ "
	}

	var unFilled string 

	for i := 0; i < widthAdj - usageBars; i++ {
		unFilled += "□ "
	}

	filled = lipgloss.NewStyle().Foreground(utils.AccentColor).Render(filled)
	unFilled = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render(unFilled)

	return lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		AlignHorizontal(lipgloss.Center).
		Render(filled + unFilled + fmt.Sprintf(" [%f%s] ", usage, "%"))
}

func retryButton() string {
	b := lipgloss.NewStyle().
		Italic(true).
		Underline(true).
		Foreground(lipgloss.Color("#d94444")).
		Render("Retry "+ "\u21BB")

	return zone.Mark("retry-button", b)
}

func filesForContextLabel(files []string) string {
	var f []string
	for _, file := range files {
		f = append(
			f, lipgloss.NewStyle().
				MarginLeft(1).
				MarginRight(1).
				Render(
					lipgloss.NewStyle().
					Foreground(utils.AccentColor).
					Bold(true).
					Render(" @") + 
					lipgloss.NewStyle().Italic(true).Render(file) + 
					" ",
				),
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, f...)
}


// -- File Input Component
type fileInputBox struct {
	ti 			     textinput.Model
	vp               viewport.Model
	matches          []utils.FileEntry
	prompt           string
}

func newFileInputBox(width, height int) fileInputBox {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "<filename> to add / rm + space to remove file"

	fib := fileInputBox {
		ti: ti,
		vp: viewport.New(width, height),
	}

	fib.ti = fib.setPrompt(fib.ti, "@")

	return fib
}

func (f *fileInputBox) setPrompt(ti textinput.Model, p string) textinput.Model {
	label := " Add file: "

	if p == "rm "{
		label = " Remove file: "
		ti.Placeholder = "<filename> to remove"
	}

	ti.Prompt = utils.BwLabelStyle.Bold(true).Render(label) + lipgloss.NewStyle().Foreground(utils.AccentColor).Bold(true).Render(" " + p)
	f.prompt = p

	return ti
}

func (f *fileInputBox) view() string {
	return lipgloss.NewStyle().MarginLeft(2).MarginBottom(1).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			f.ti.View(),
			f.vp.View(),
		),
	)
}

// Slash Command Input Componenet
type slashCmdCtrl struct {
	ti 			     textinput.Model
	vp               viewport.Model
	matches          []string
	topMatch         slashCommand
	prompt           string
	acceptingCmd     bool
	openCmd          string
}

func newSlashCmdCtrl(width, height int) slashCmdCtrl {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "/<command>"
	ti.Prompt = utils.BwLabelStyle.Bold(true).Render(" Command: ") + " "

	return slashCmdCtrl {
		ti: ti,
		vp: viewport.New(width, height),
	}
}

func (s *slashCmdCtrl) updateMatches() {
	s.matches = utils.ListCmdMatches(s.ti.Value(), slashCmdNames)
	s.vp.SetContent(strings.Join(s.matches, "\n"))
	s.topMatch = slashCommand{}

	if len(s.matches) > 0 {
		for _, slashCmd := range slashCommands {
			if slashCmd.name == s.matches[0] {
				s.topMatch = slashCmd
				break
			}
		}
	}
}

func (s *slashCmdCtrl) view() string {
	return lipgloss.NewStyle().MarginLeft(2).MarginBottom(1).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			s.ti.View(),
			s.vp.View(),
		),
	)
}