package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui"
)

func main() {
	zone.NewGlobal()
	zone.SetEnabled(true)
	config.Init()

	model := ui.InitialModel()
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, runErr := program.Run()
	closeErr := model.Close()
	if runErr != nil {
		fmt.Println("error: ", runErr)
	}
	if closeErr != nil {
		fmt.Println("error: ", closeErr)
	}
	if runErr != nil || closeErr != nil {
		os.Exit(1)
	}
}
