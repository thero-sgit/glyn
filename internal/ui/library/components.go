package library

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
)

const glynSmallText = `
███ █   █ █ █ █
█   █    █  ███
█ █ █    █  ███
███ ███  █  █ █`

var (
	SendPromptBtnStates = []string{
		HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render("send " + "\u27A4"),
		BwLabelStyle.PaddingLeft(1).PaddingRight(1).Render("stop " + "\u25A0"),
	}
	SendPromptBtnState = SendPromptBtnStates[0]
)

func Burner(width int) string {
	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(AccentColor).
		Render(glynSmallText)

	model := lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(SubtleStyle.Render("\u27E1 " + config.Config.Model.Name))

	return lipgloss.NewStyle().
		Width(width).
		AlignHorizontal(lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center, logo, model),
		)
}

func CreateTextArea(width, height int, value string) textarea.Model {
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

func LabelValueBand(label, value, secondaryValue string) string {
	label = HighlighedStyle.
		PaddingLeft(1).
		Render(label)

	value = HighlighedStyle.
		PaddingLeft(1).
		PaddingRight(1).
		Render(value)

	if secondaryValue != "" {
		secondaryValue = lipgloss.NewStyle().
			Background(AccentColor).
			Foreground(lipgloss.Color("#282828")).
			PaddingRight(1).
			Render(secondaryValue)
	}

	joined := lipgloss.JoinHorizontal(lipgloss.Left, label, value, secondaryValue)

	output := BwLabelStyle.
		AlignHorizontal(lipgloss.Left).
		Render(joined)

	return output
}

func TokenUsageProgessBar(usage float32) string {
	widthAdj := 25
	usageBars := int(usage * (float32(widthAdj) / 100))

	var filled string

	for i := 0; i < usageBars; i++ {
		filled += "■ "
	}

	var unFilled string

	for i := 0; i < widthAdj-usageBars; i++ {
		unFilled += "□ "
	}

	filled = lipgloss.NewStyle().Foreground(AccentColor).Render(filled)
	unFilled = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render(unFilled)

	return lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Render(filled + unFilled + fmt.Sprintf(" [%f%s] ", usage, "%"))
}

func RetryButton(zoneID string) string {
	b := lipgloss.NewStyle().
		Italic(true).
		Underline(true).
		Foreground(lipgloss.Color("#d94444")).
		Render("Retry " + "\u21BB")

	return zone.Mark(zoneID, b)
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
						Render(" @")+
						lipgloss.NewStyle().Italic(true).Render(file)+
						" ",
				),
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, f...)
}
