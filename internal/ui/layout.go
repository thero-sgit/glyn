package ui

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
)

func (m Model) headerBar(headerHeight int) string {
	version := subtleStyle.Render("v0.4.2")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(headerHeight).
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Right).
		PaddingRight(2).
		Render(version)
}


func (m Model) chatUi(middleHeight int) string {
	taHeight := 2
	if m.textarea.Length() > 0 {
		taHeight = 6
	}
	m.textarea.SetHeight(taHeight)
	m.viewport.Height = middleHeight - (taHeight + 1)

	prompt := lipgloss.NewStyle().
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("  >  ")

	inputBoxContent := lipgloss.NewStyle().
		Render(lipgloss.JoinHorizontal(lipgloss.Top, prompt, m.textarea.View()))

	inputBoxContent = zone.Mark("text-area", inputBoxContent)

	boxStyle := lipgloss.NewStyle().
		Width(m.width-6)

	workingDirAndSessionLabel := lipgloss.NewStyle().
		Render(
			lipgloss.JoinHorizontal(lipgloss.Left, dirAndSessionLabel(config.Config.SanitizedWd, m.chatCentre.sessionName)),
		)

	addContext := lipgloss.NewStyle().
		MarginLeft(1).
		PaddingLeft(1).
		PaddingRight(1).
		Render("\uFF0B")

	addContextbutton := zone.Mark("add-context-btn", addContext)
	modeLabelButton  := zone.Mark("mode-toggle-btn", labelValueBand(config.Config.Mode.Icon, config.Config.Mode.Name))

	inpLabels := lipgloss.NewStyle().PaddingLeft(1).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			addContextbutton, 
			modeLabelButton + " ",
			workingDirAndSessionLabel,
		),
	)
	
	promptBox := boxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, inputBoxContent, inpLabels),
	)

	activity := lipgloss.NewStyle().
		Width(m.width-4).
		Height(middleHeight - (taHeight + 1)).
		MarginLeft(2).
		MarginRight(2).
		Render(m.viewport.View())

	main := lipgloss.JoinVertical(lipgloss.Top, activity, promptBox)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(middleHeight).
		PaddingLeft(2).
		PaddingRight(2).
		Render(main)
}

func (m Model) footerBar(footerHeight int) string {
	joined := subtleStyle.Render("ctrl+e to write prompt   ? for help   esc/ctrl+c to quit")

	return lipgloss.NewStyle().
		PaddingRight(2).
		PaddingLeft(2).
		Width(m.width).
		AlignHorizontal(lipgloss.Right).
		Height(footerHeight).
		Render(joined)
}

func chunkString(s string, n int) []string {
    var chunks []string
    for i := 0; i < len(s); i += n {
        end := i + n
        if end > len(s) {
            end = len(s)
        }
        chunks = append(chunks, s[i:end])
    }
    return chunks
}