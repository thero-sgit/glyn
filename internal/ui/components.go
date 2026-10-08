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
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlighedStyle = lipgloss.NewStyle().Background(accentColor).Foreground(lipgloss.Color("#000000"))
	bwLabelStyle = lipgloss.NewStyle().Background(lipgloss.Color("#f0efef")).Foreground(lipgloss.Color("#121111"))
	accentColor  = lipgloss.Color("#D4A574")
	boldStyle    = lipgloss.NewStyle().Bold(true)
	
	helpCmdTabState   = 0
	helpCmdGeneralTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				boldStyle.Render("esc") + "/" + boldStyle.Render("ctrl+c") + " to quit xyn",
				boldStyle.Render("/") + " display '/' command input screen",

				subtleStyle.MarginTop(1).Render("AI interaction & config"),
				boldStyle.Render("alt+enter") + "/click " + sendPromptBtnStates[0] + " send prompt",
				boldStyle.Render("@") + "/click " + boldStyle.Render("@") + " to add files",
				"click " + labelValueBand("\u27E1", "`model-id`") + " to view models list",
				"click " + labelValueBand("?/" + "\u2726", "`mode`") + " to switch mode",
				
				subtleStyle.MarginTop(1).Render("In chat screen"),
				boldStyle.Render("pgup, pgdn, up, down arrows") + "/" + boldStyle.Render("mouse") + " to scroll",
				boldStyle.Render("ctrl+e") + " to focus prompt text input",
				boldStyle.Render("ctrl+x") + "/click " + sendPromptBtnStates[1] + " to interupt session",
			),
		)

	helpCmdCommandsTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				boldStyle.Render("/help") + " display helpful keyboard shortcuts",
				boldStyle.Render("/status") + " display usage and general status",
				boldStyle.Render("/model") + " display models list",
			),
		)

	slashCommands = []slashCommand {
		slashCommand{ name: "/help", pretty: "/help  ", comp: helpCmdComponent},
		slashCommand{ name: "/status", pretty: "/status", comp: statusCmdComponent},
		slashCommand{ name: "/model", pretty: "/model"},
	}
	slashCmdNames = func() []string {
		var names []string
		for _, slashCmd := range slashCommands { names = append(names, slashCmd.name)}
		return names
	}()

	sendPromptBtnStates = []string {
		highlighedStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("send " + "\u27A4"),
		bwLabelStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("stop " + "\u25A0"),
	}

	sendPromptState = sendPromptBtnStates[0]
)

func bunner() string {
	inf := lipgloss.NewStyle().MarginLeft(2).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			boldStyle.Render("xyn ") + subtleStyle.Bold(true).Render(config.Config.Version),
			boldStyle.Foreground(accentColor).Render("\u27E1 ") + subtleStyle.Render(config.Config.Model),
			boldStyle.Foreground(accentColor).Render("@ ") + subtleStyle.Underline(true).Render("console.groq.com"),
			boldStyle.Foreground(accentColor).Render("~") + subtleStyle.Render(config.Config.WorkingDir),
		),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(accentColor).Render(xynLargeText), subtleStyle.Render(codeLargeText)),
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
		generalTabBtn  = highlighedStyle.PaddingLeft(1).PaddingRight(1).Render("General")
		commandsTabBtn = subtleStyle.Render(" Commands ")
		content 	   = helpCmdGeneralTabContent
	} else {
		generalTabBtn  = subtleStyle.Render(" General ")
		commandsTabBtn = highlighedStyle.PaddingLeft(1).PaddingRight(1).Render("Commands")
		content        = helpCmdCommandsTabContent
	}

	generalTabBtn  = zone.Mark("helpCmd-general-btn", generalTabBtn)
	commandsTabBtn = zone.Mark("helpCmd-commands-btn", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	headerBar := lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinHorizontal(
			lipgloss.Left, 
			bwLabelStyle.Render(" /help "),
			" ",
			tabs,
			" ",
			"Useful Shortcuts & '/' Commands",
			" ",
			subtleStyle.MarginLeft(5).Render("press 'esc' to close   left/right arrows to switch tabs"),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		headerBar,
		content,
	)
}

func statusCmdComponent() string {
	cmdLabel := bwLabelStyle.Render(" /status ")

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
			boldStyle.Render(subtleStyle.Render("\u27E1")) +
			" " + modelInfo.Name + 
			" (" + modelInfo.Id + ") " + 
			subtleStyle.Render(modelInfo.OwnedBy),
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
				subtleStyle.Render(val + " ") + s + "  ",
			)
		}
	}

	capabilities := lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			subtleStyle.Render("Capabilities"),
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
		Render(lipgloss.NewStyle().Foreground(accentColor).Render("~") + subtleStyle.Faint(true).Render(directoryPath))
	
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
	label = highlighedStyle.
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Render(label )

	value = highlighedStyle.
		Bold(true).
		PaddingRight(1).
		Render(value)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, label, value)

	output := bwLabelStyle.
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

	filled = lipgloss.NewStyle().Foreground(accentColor).Render(filled)
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
					Foreground(accentColor).
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

	ti.Prompt = bwLabelStyle.Bold(true).Render(label) + lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(" " + p)
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
	ti.Prompt = bwLabelStyle.Bold(true).Render(" Command: ") + " "

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