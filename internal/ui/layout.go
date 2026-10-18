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

func (m *Model) inputBox(h int) string {
	taHeight := 3
	if m.textarea.Length() > 0 || m.chatCentre.fileSelector.Focused || m.slashCmdCon.Focused {
		taHeight = 7
	}

	if m.chatCentre.fileSelector.Focused {
		m.chatCentre.fileSelector.Viewport.Height = taHeight - 2
		m.chatCentre.fileSelector.Viewport.Width = m.width - 8

	} else if m.slashCmdCon.Focused {
		m.slashCmdCon.Viewport.Height = taHeight - 2
		m.slashCmdCon.Viewport.Width = m.width - 8

	} else {
		m.textarea.SetHeight(taHeight - 1)
	}
	taHeight++

	m.mHeight = h - taHeight

	//
	prompt := lipgloss.NewStyle().
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("  >  ")

	textArea := lipgloss.JoinHorizontal(lipgloss.Left, prompt, m.textarea.View())
	textArea = zone.Mark("text-area", textArea)
	sendStopBtn := zone.Mark("send-prompt-btn", stylesandcomponents.SendPromptState)

	var inputBoxContent string
	if m.chatCentre.fileSelector.Focused {
		inputBoxContent = m.chatCentre.fileSelector.View()
	} else if m.slashCmdCon.Focused {
		inputBoxContent = m.slashCmdCon.View()
	} else {
		inputBoxContent = lipgloss.JoinVertical(lipgloss.Top, stylesandcomponents.FilesForContextLabel(m.chatCentre.filesForContext), lipgloss.NewStyle().Render(textArea, sendStopBtn))
	}

	boxStyle := lipgloss.NewStyle().
		Width(m.width - 6)

	addFile := lipgloss.NewStyle().
		MarginLeft(1).
		PaddingLeft(1).
		PaddingRight(2).
		Bold(true).
		Render("@")

	addFileButton := zone.Mark("add-file-btn", addFile)
	modeLabelButton := zone.Mark("mode-toggle-btn", stylesandcomponents.LabelValueBand(config.Config.Mode.Icon, config.Config.Mode.Name, ""))
	modelLabelBtn := zone.Mark("model-selector-btn", stylesandcomponents.LabelValueBand("\u27E1", config.Config.Model.Name, config.Config.Effort.Name))

	inpLabels := lipgloss.NewStyle().PaddingLeft(1).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			addFileButton,
			modelLabelBtn+" ",
			modeLabelButton+" ",
		),
	)

	promptBox := boxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top, inputBoxContent, inpLabels),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(taHeight).
		PaddingLeft(2).
		PaddingRight(2).
		Render(promptBox)
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
