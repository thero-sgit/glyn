package ui

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

func (m Model) headerBar(headerHeight int) string {
	version := stylesandcomponents.SubtleStyle.Render(config.Config.Version)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(headerHeight).
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Right).
		PaddingRight(2).
		Render(version)
}


func (m Model) chatUi(middleHeight int) string {
	taHeight := 3
	if m.textarea.Length() > 0 || m.chatCentre.fileInpBoxOpen || m.slashCmdCon.Focused {
		taHeight = 7
	}

	if m.chatCentre.fileInpBoxOpen {
		m.chatCentre.fileInpBox.Vp.Height = taHeight - 2
		m.chatCentre.fileInpBox.Vp.Width  = m.width - 8

	} else if m.slashCmdCon.Focused {
		m.slashCmdCon.Viewport.Height = taHeight - 1
		m.slashCmdCon.Viewport.Width  = m.width - 8
		
	} else {
		m.textarea.SetHeight(taHeight-1)
	}

	m.viewport.Height = middleHeight - (taHeight + 1)

	prompt := lipgloss.NewStyle().
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("  >  ")

	textArea := lipgloss.JoinHorizontal(lipgloss.Left, prompt, m.textarea.View())
	textArea = zone.Mark("text-area", textArea)
	sendStopBtn := zone.Mark("send-prompt-btn", stylesandcomponents.SendPromptState)

	var inputBoxContent string
	if m.chatCentre.fileInpBoxOpen {
		inputBoxContent = m.chatCentre.fileInpBox.View()
	} else if m.slashCmdCon.Focused {
		inputBoxContent = m.slashCmdCon.View()
	} else {
		inputBoxContent = lipgloss.JoinVertical(lipgloss.Top, stylesandcomponents.FilesForContextLabel(m.chatCentre.filesForContext), lipgloss.NewStyle().Render(textArea, sendStopBtn))
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
	modeLabelButton  := zone.Mark("mode-toggle-btn", stylesandcomponents.LabelValueBand(config.Config.Mode.Icon, config.Config.Mode.Name))
	modelLabelBtn    := zone.Mark("model-selector-btn", stylesandcomponents.LabelValueBand("\u27E1", config.Config.Model))

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
	joined := stylesandcomponents.SubtleStyle.Render("ctrl+e to write prompt   ? for help   esc/ctrl+c to quit")

	return lipgloss.NewStyle().
		PaddingRight(2).
		PaddingLeft(2).
		Width(m.width).
		AlignHorizontal(lipgloss.Right).
		Height(footerHeight).
		Render(joined)
}