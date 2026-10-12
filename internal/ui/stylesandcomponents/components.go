package stylesandcomponents

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ai"
	"github.com/thero-sgit/xyn/internal/common"
	"github.com/thero-sgit/xyn/internal/config"
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
	SendPromptBtnStates = []string {
		HighlighedStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("send " + "\u27A4"),
		BwLabelStyle.PaddingLeft(1).PaddingRight(1).Bold(true).Render("stop " + "\u25A0"),
	}
	SendPromptState = SendPromptBtnStates[0]
)

func Bunner() string {
	inf := lipgloss.NewStyle().MarginLeft(2).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			BoldStyle.Render("xyn ") + SubtleStyle.Bold(true).Render(config.Config.Version),
			BoldStyle.Foreground(AccentColor).Render("\u27E1 ") + SubtleStyle.Render(config.Config.Model),
			BoldStyle.Foreground(AccentColor).Render("@ ") + SubtleStyle.Underline(true).Render("console.groq.com"),
			BoldStyle.Foreground(AccentColor).Render("~") + SubtleStyle.Render(config.Config.WorkingDir),
		),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(AccentColor).Render(xynLargeText), SubtleStyle.Render(codeLargeText)),
		inf,
	)
}

func statusCmdComponent() string {
	cmdLabel := BwLabelStyle.Render(" /status ")

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
			BoldStyle.Render(SubtleStyle.Render("\u27E1")) +
			" " + modelInfo.Name + 
			" (" + modelInfo.Id + ") " + 
			SubtleStyle.Render(modelInfo.OwnedBy),
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
				SubtleStyle.Render(val + " ") + s + "  ",
			)
		}
	}

	capabilities := lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			SubtleStyle.Render("Capabilities"),
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

func CreateTextArea(width, height int, value string,) textarea.Model {
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
		Render(lipgloss.NewStyle().Foreground(AccentColor).Render("~") + SubtleStyle.Faint(true).Render(directoryPath))
	
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

func LabelValueBand(label, value string) string {
	label = HighlighedStyle.
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Render(label )

	value = HighlighedStyle.
		Bold(true).
		PaddingRight(1).
		Render(value)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, label, value)

	output := BwLabelStyle.
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

	filled = lipgloss.NewStyle().Foreground(AccentColor).Render(filled)
	unFilled = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render(unFilled)

	return lipgloss.NewStyle().
		MarginTop(1).
		MarginBottom(1).
		AlignHorizontal(lipgloss.Center).
		Render(filled + unFilled + fmt.Sprintf(" [%f%s] ", usage, "%"))
}

func RetryButton() string {
	b := lipgloss.NewStyle().
		Italic(true).
		Underline(true).
		Foreground(lipgloss.Color("#d94444")).
		Render("Retry "+ "\u21BB")

	return zone.Mark("retry-button", b)
}

func FilesForContextLabel(files []string) string {
	var f []string
	for _, file := range files {
		f = append(
			f, lipgloss.NewStyle().
				MarginLeft(1).
				MarginRight(1).
				Render(
					lipgloss.NewStyle().
					Foreground(AccentColor).
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
type FileInputBox struct {
	Ti 			     textinput.Model
	Vp               viewport.Model
	Matches          []common.FileEntry
	Prompt           string
}

func NewFileInputBox(width, height int) FileInputBox {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "<filename> to add / rm + space to remove file"

	fib := FileInputBox {
		Ti: ti,
		Vp: viewport.New(width, height),
	}

	fib.Ti = fib.SetPrompt(fib.Ti, "@")

	return fib
}

func (f *FileInputBox) SetPrompt(ti textinput.Model, p string) textinput.Model {
	label := " Add file: "

	if p == "rm "{
		label = " Remove file: "
		ti.Placeholder = "<filename> to remove"
	}

	ti.Prompt = BwLabelStyle.Bold(true).Render(label) + lipgloss.NewStyle().Foreground(AccentColor).Bold(true).Render(" " + p)
	f.Prompt = p

	return ti
}

func (f *FileInputBox) View() string {
	return lipgloss.NewStyle().MarginLeft(2).MarginBottom(1).Render(lipgloss.JoinVertical(
			lipgloss.Top,
			f.Ti.View(),
			f.Vp.View(),
		),
	)
}