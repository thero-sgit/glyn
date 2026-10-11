package submodels

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/common"
)

var (
	helpCmdTabState   = 0
	helpCmdGeneralTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				common.BoldStyle.Render("esc") + "/" + common.BoldStyle.Render("ctrl+c") + " to quit xyn",
				common.BoldStyle.Render("/") + " display '/' command input screen",

				common.SubtleStyle.MarginTop(1).Render("AI interaction & config"),
				common.BoldStyle.Render("alt+enter") + "/click " + sendPromptBtnStates[0] + " send prompt",
				common.BoldStyle.Render("@") + "/click " + common.BoldStyle.Render("@") + " to add files",
				"click " + labelValueBand("\u27E1", "`model-id`") + " to view models list",
				"click " + labelValueBand("?/" + "\u2726", "`mode`") + " to switch mode",
				
				common.SubtleStyle.MarginTop(1).Render("In chat screen"),
				common.BoldStyle.Render("pgup, pgdn, up, down arrows") + "/" + common.BoldStyle.Render("mouse") + " to scroll",
				common.BoldStyle.Render("ctrl+e") + " to focus prompt text input",
				common.BoldStyle.Render("ctrl+x") + "/click " + sendPromptBtnStates[1] + " to interupt session",
			),
		)

	helpCmdCommandsTabContent = lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				common.BoldStyle.Render("/help") + " display helpful keyboard shortcuts",
				common.BoldStyle.Render("/status") + " display usage and general status",
				common.BoldStyle.Render("/model") + " display models list",
			),
		)
)