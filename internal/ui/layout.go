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
	if m.textarea.Length() > 0 || m.chatCentre.fileInpBoxOpen {
		taHeight = 6
	}

	if m.chatCentre.fileInpBoxOpen {
		if m.chatCentre.fileInpBox.vp.Height == 0 {
			m.chatCentre.fileInpBox.vp.Height = taHeight - 2
			m.chatCentre.fileInpBox.vp.Width  = m.width - 8
		}		
	} else {
		m.textarea.SetHeight(taHeight)
	}

	m.viewport.Height = middleHeight - (taHeight + 1)

	prompt := lipgloss.NewStyle().
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("  >  ")

	textArea := lipgloss.JoinHorizontal(lipgloss.Left, prompt, m.textarea.View())
	textArea = zone.Mark("text-area", textArea)
	sendStopBtn := zone.Mark("send-prompt-btn", sendPromptState)

	var inputBoxContent string
	if m.chatCentre.fileInpBoxOpen {
		inputBoxContent = m.chatCentre.fileInpBox.view()
	} else {
		inputBoxContent = lipgloss.NewStyle().Render(textArea, sendStopBtn)
	}	

	boxStyle := lipgloss.NewStyle().
		Width(m.width-6)

	addFile := lipgloss.NewStyle().
		MarginLeft(1).
		PaddingLeft(1).
		PaddingRight(2).
		Bold(true).
		Render("@")

	addFileButton := zone.Mark("add-file-btn", addFile)
	modeLabelButton  := zone.Mark("mode-toggle-btn", labelValueBand(config.Config.Mode.Icon, config.Config.Mode.Name))
	modelLabelBtn    := zone.Mark("model-selector-btn", labelValueBand("\u27E1", config.Config.Model))

	inpLabels := lipgloss.NewStyle().PaddingLeft(1).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			addFileButton,
			modelLabelBtn + " ",
			modeLabelButton + " ",
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