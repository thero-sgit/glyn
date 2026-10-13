package submodels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

type LLModels struct {
	Focused  bool
	modeldd  Dropdown
	effortdd Dropdown
}

func (l *LLModels) Init() tea.Cmd {
	return nil
}

func NewModelsModel() LLModels {
	return LLModels{
		modeldd:  NewDropdown("models", config.Config.Model, []string{"model-1", "model-2"}),
		effortdd: NewDropdown("effort", "medium", []string{"low", "medium", "high"}),
	}
}

func (l *LLModels) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			l.modeldd.Focused  = false
			l.effortdd.Focused = false

			switch {
			case zone.Get("models-dropdown-menu").InBounds(msg):
				l.modeldd.Focused  = true
				l.effortdd.Focused = false

				updatedModel, cmd := l.modeldd.Update(msg)
				l.modeldd = *updatedModel.(*Dropdown)
				return l, cmd

			case zone.Get("effort-dropdown-menu").InBounds(msg):
				l.modeldd.Focused  = false
				l.effortdd.Focused = true

				updatedModel, cmd := l.effortdd.Update(msg)
				l.effortdd = *updatedModel.(*Dropdown)
				return l, cmd
			}
		}

		return l, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			l.updateEither(msg)

			config.Config.Model  = l.modeldd.Selected
			config.Config.Effort = l.effortdd.Selected

		case "up", "down":
			return l.updateEither(msg)			
		}
	}

	return l, cmd
}

func (l *LLModels) updateEither(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case l.modeldd.Focused:
		updatedModel, cmd := l.modeldd.Update(msg)
		l.modeldd = *updatedModel.(*Dropdown)
		return l, cmd

	case l.effortdd.Focused:
		updatedModel, cmd := l.effortdd.Update(msg)
		l.effortdd = *updatedModel.(*Dropdown)
		return l, cmd
	}

	return l, nil
}

func (l *LLModels) View() string {
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		stylesandcomponents.BwLabelStyle.Render(" /model "),
		" Model: ",
		l.modeldd.View(),
		" Effort ",
		l.effortdd.View(),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
	)
}
