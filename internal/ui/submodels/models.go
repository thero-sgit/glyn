package submodels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ai"
	"github.com/thero-sgit/xyn/internal/config"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

type LLModels struct {
	Focused bool
	tab     int
}

func (l *LLModels) Init() tea.Cmd {
	return nil
}

func NewModelsModel() LLModels {
	return LLModels{}
}

func (l *LLModels) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			case zone.Get("model-toggle").InBounds(msg):
				if l.tab == 0 {
					config.Config.Model.Toggle()
				} else {
					l.tab = 0
				}
				
				return l, nil

			case zone.Get("effort-toggle").InBounds(msg):
				if l.tab == 1 {
					config.Config.Effort.Toggle()
				} else {
					l.tab = 1
				}

				return l, nil
			}
		}

		return l, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if l.tab == 0{
				config.Config.Model.Toggle()
			} else {
				config.Config.Effort.Toggle()
			}

		case "left":
			l.tab = 0

		case "right":
			l.tab = 1
		}
	}

	return l, nil
}

func (l *LLModels) View() string {
	var generalTabBtn string
	var commandsTabBtn string

	if l.tab == 0 {
		generalTabBtn = stylesandcomponents.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render(config.Config.Model.Pretty())
		commandsTabBtn = stylesandcomponents.SubtleStyle.Render(config.Config.Effort.Pretty())
	} else {
		generalTabBtn = stylesandcomponents.SubtleStyle.Render(config.Config.Model.Pretty())
		commandsTabBtn = stylesandcomponents.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render(config.Config.Effort.Pretty())
	}

	generalTabBtn = zone.Mark("model-toggle", generalTabBtn)
	commandsTabBtn = zone.Mark("effort-toggle", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	header := lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		stylesandcomponents.BwLabelStyle.MarginRight(1).Render(" /model "),
		tabs,
		"  Compatable models from Groq (console.groq.com/docs/models)",
		stylesandcomponents.SubtleStyle.MarginLeft(5).Render("press 'esc' to close"),
		),
	)

	models := []ai.GroqModelInfo {
		{
			Id:   "model-1",
			Name: "Model-One",
			OwnedBy: "Baseby",
			SupportedFeatures: []string{"tools", "reasoning"},
		},
	}

	modelcards := []string{
		modelCard(models[0]),
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Left, modelcards...),
	)
}


func modelCard(modelInfo ai.GroqModelInfo) string {
	header := lipgloss.NewStyle().Render(
		stylesandcomponents.BoldStyle.Render(stylesandcomponents.SubtleStyle.Render("\u27E1")) +
			" " + modelInfo.Name +
			" (" + modelInfo.Id + ") " +
			stylesandcomponents.SubtleStyle.Render(modelInfo.OwnedBy),
	)

	c := map[string]string{
		"tools":     "\u2692",
		"reasoning": "\u223f",
	}

	var labels []string
	for _, s := range modelInfo.SupportedFeatures {
		val, ok := c[s]
		if ok {
			labels = append(
				labels,
				stylesandcomponents.SubtleStyle.Render(val+" ")+s+"  ",
			)
		}
	}

	capabilities := lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			stylesandcomponents.SubtleStyle.Render("Capabilities"),
			lipgloss.JoinHorizontal(lipgloss.Left, labels...),
		),
	)

	card := lipgloss.NewStyle().Margin(0, 1, 1, 0).Padding(0, 1, 0, 1).Border(lipgloss.RoundedBorder()).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			capabilities,
		),
	)

	return card
}