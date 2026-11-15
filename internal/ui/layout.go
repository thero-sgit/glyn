package ui

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

func (m Model) headerBar(headerHeight int) string {
	version := library.SubtleStyle.Render(config.Config.Version)
	if m.modelsLoading {
		version = lipgloss.JoinHorizontal(
			lipgloss.Right,
			m.modelsSpinner.View(),
			" ",
			version,
		)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(headerHeight).
		AlignVertical(lipgloss.Bottom).
		AlignHorizontal(lipgloss.Right).
		PaddingRight(2).
		Render(version)
}

func (m *Model) inputBox(h int) string {
	taHeight := 4
	if m.textarea.Length() > 0 || m.fileSelector.Focused || m.slashCmdCon.Focused {
		taHeight = 8
	}

	if m.fileSelector.Focused {
		m.fileSelector.Viewport.Height = taHeight - 3
		m.fileSelector.Viewport.Width = m.width - 8

	} else if m.slashCmdCon.Focused {
		m.slashCmdCon.Viewport.Height = taHeight - 3
		m.slashCmdCon.Viewport.Width = m.width - 8

	} else {
		m.textarea.SetHeight(taHeight - 1)
	}

	m.mHeight = h - taHeight

	//
	prompt := lipgloss.NewStyle().
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("  >  ")

	textArea := lipgloss.JoinHorizontal(lipgloss.Left, prompt, m.textarea.View())
	textArea = zone.Mark("text-area", textArea)

	var inputBoxContent string
	if m.fileSelector.Focused {
		inputBoxContent = m.fileSelector.View()
	} else if m.slashCmdCon.Focused {
		inputBoxContent = m.slashCmdCon.View()
	} else {
		inputBoxContent = textArea
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
	modeLabelButton := zone.Mark("mode-toggle-btn", library.LabelValueBand(config.Config.Mode.Icon, config.Config.Mode.Name, ""))
	modelLabelBtn := zone.Mark("model-selector-btn", library.LabelValueBand("\u27E1", config.Config.Model.Name, config.Config.Effort.Name))

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
	joined := library.SubtleStyle.Render("ctrl+e to write prompt   ? for help   ctrl+c to quit")

	return lipgloss.NewStyle().
		PaddingRight(2).
		PaddingLeft(2).
		Width(m.width).
		AlignHorizontal(lipgloss.Right).
		Height(footerHeight).
		Render(joined)
}
