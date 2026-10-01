package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ai"
	"github.com/thero-sgit/xyn/internal/config"
)

const xynLargeText = `
██╗ ██╗██╗ ██╗████╗██╗
 ████╔╝██████║██║█║██║
██║ ██╗    ██║██║████║
╚═╝ ╚═╝    ╚═╝╚═╝╚═══╝
`

const codeLargeText = `
█████╗██████╗█████╗ ██████╗
██╔══╝██║ ██║██║ ██╗████╔═╝
█████╗██████║█████╔╝██████╗
╚════╝╚═════╝╚════╝ ╚═════╝
`
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
				"click " + boldStyle.Render("\uFF0B") + " to add context",
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
			),
		)

	slashCommands = []uiOption {
		&slashCommand{ name: "/help", pretty: "/help  ", comp: helpCmdComponent},
		&slashCommand{ name: "/status", pretty: "/status", comp: statusCmdComponent},
		&slashCommand{ name: "/models", pretty: "/models"},
	}

	sendPromptBtnStates = []string {
		highlighedStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("send " + "\u27A4"),
		bwLabelStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("stop " + "\u25A0"),
	}

	sendPromptState = sendPromptBtnStates[0]
)

func bunner() string {
	inf := lipgloss.NewStyle().MarginLeft(3).MarginTop(1).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			boldStyle.Render("xyn ") + subtleStyle.Bold(true).Render("v0.4.2"),
			boldStyle.Foreground(accentColor).Render("\u27E1 ") + subtleStyle.Render(config.Config.Model),
			boldStyle.Foreground(accentColor).Render("~") + subtleStyle.Render(config.Config.WorkingDir),
		),
	)

	largeText := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(accentColor).Render(xynLargeText),
		subtleStyle.Render(codeLargeText),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		largeText,
		lipgloss.NewStyle().MarginBottom(1).Render(inf),
	)
}

type uiOption interface {
	setPretty(string)
	getPretty() string
	getName()   string
	component() string
}

type slashCommand struct{
	name   string
	pretty string
	comp   func() string
}

func (s *slashCommand) setPretty(pretty string) {
	s.pretty = pretty
}

func (s *slashCommand) getPretty() string {
	return s.pretty
}

func (s *slashCommand) getName() string {
	return s.name
}

func (s *slashCommand) component() string {
	return s.comp()
}

type selector interface {
	textArea()      textinput.Model
	setTaWSlash()
	setTaPH(string)

	opts() 	 		[]uiOption
	setOpts([]uiOption)
	optsLen() 	 	int
	setOptsLen(int)

	hiIdx()        	int
	setHiIdx(int)
}

type slashCmdCtrl struct {
	ta               textinput.Model
	options          []uiOption
	cmdLen           int
	acceptingCmd     bool
	openCmd          string
	highlightedIndex int	 
}

func (s *slashCmdCtrl) textArea() textinput.Model {
	return s.ta
}

func (s *slashCmdCtrl) setTaWSlash() {
	s.ta.SetValue("/" + s.textArea().Value())
	s.ta.CursorEnd()
}

func (s *slashCmdCtrl) setTaPH(ph string) {
	s.ta.Placeholder = ph
}

func (s *slashCmdCtrl) opts() []uiOption {
	return s.options
}

func (s *slashCmdCtrl) setOpts(opts []uiOption) {
	s.options = opts
}

func (s *slashCmdCtrl) optsLen() int {
	return s.cmdLen
}

func (s *slashCmdCtrl) setOptsLen(l int) {
	s.cmdLen = l
}

func (s *slashCmdCtrl) hiIdx() int {
	return s.highlightedIndex
}

func (s *slashCmdCtrl) setHiIdx(i int) {
	s.highlightedIndex = i
}

func newSlashCmdCtrl() slashCmdCtrl {
	ta := textinput.New()
	ta.Focus()
	ta.CharLimit = 15

	return slashCmdCtrl{
		ta:               ta,
		cmdLen:           len(slashCommands),
		acceptingCmd:     true,
		options: slashCommands,
		highlightedIndex: 0,
	}
}

func viewSelector[T selector](s T) string {
	if len(s.textArea().Value()) > 0 && !strings.HasPrefix(s.textArea().Value(), "/") { s.setTaWSlash() }

	opts := []uiOption{}
	for _, option := range s.opts() {
		if strings.HasPrefix(option.getName(), s.textArea().Value()) {
			opts = append(opts, option)
		}
	} // FIX OPTS BUG!
	
	s.setOpts(opts)
	s.setOptsLen(len(opts))

	toDisplay := func() []string {
		var r []string
		for i, opt := range opts {
			p := opt.getPretty()
			if i == s.hiIdx() {
				p = lipgloss.NewStyle().Foreground(accentColor).Render("> ") + opts[s.hiIdx()].getPretty()
			}
			r = append(r, p)
		}
		return  r
	}()

	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(27).MarginRight(5).Render("command: " + s.textArea().View()),
				subtleStyle.Render("press 'esc' to close"),
			),
			strings.Join(toDisplay, "\n"),
		),
	)
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

func modelCard(modelInfo ai.GroqModelInfo, width int) string {
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

func modelsCmdComponent(content string, models []ai.GroqModelInfo, width int) string {
	cmdLabel := lipgloss.JoinHorizontal(
		lipgloss.Left,
		bwLabelStyle.Render(" /models "),
		" Compatible models from Groq ",
		subtleStyle.Render("(console.groq.com/docs/models)"),
		subtleStyle.MarginLeft(5).Render("press 'esc' to close"),
	)

	if content == "" && len(models)>0 {
		for _, model := range models {
			content += modelCard(model, width) + "\n"
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		cmdLabel,
		lipgloss.NewStyle().MarginTop(1).Render(content),
	)
}

func createTextArea(width, height int, value string,) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = ` Try "how does <filename> work?"`
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	
	ta.SetWidth(width - 11)
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