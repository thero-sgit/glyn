package library

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	SubtleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	HighlighedStyle = lipgloss.NewStyle().Background(AccentColor).Foreground(lipgloss.Color("#000000"))
	BwLabelStyle    = lipgloss.NewStyle().Background(lipgloss.Color("#f0efef")).Foreground(lipgloss.Color("#121111"))
	AccentColor     = lipgloss.Color("#dfdedc")
	BoldStyle       = lipgloss.NewStyle().Bold(true)
)

func ColorForOffset(offset float64) lipgloss.Style {
	level := 235 + int((offset+1)/2*23)
	return lipgloss.NewStyle().Foreground(
		lipgloss.Color(fmt.Sprintf("%d", level)),
	)
}
