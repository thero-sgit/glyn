package submodels

import (
	"cmp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ai"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/api"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type LLModels struct {
	viewport viewport.Model
	tab      int
	models   []ai.GroqModelInfo
}

func (l *LLModels) Init() tea.Cmd {
	return nil
}

func NewModelsModel() tea.Model {
	return &LLModels{
		viewport: viewport.New(0, 0),
	}
}

func (l *LLModels) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.viewport.Width, l.viewport.Height = msg.Width, msg.Height
		return l, nil

	case tea.MouseMsg:
		var cmd tea.Cmd
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch {
			case zone.Get("model-toggle").InBounds(msg):
				if l.tab == 0 {
					config.Config.Model.Toggle()
				} else {
					l.tab = 0
				}
				
				return l, cmd

			case zone.Get("effort-toggle").InBounds(msg):
				if l.tab == 1 {
					config.Config.Effort.Toggle()
				} else {
					l.tab = 1
				}

				return l, cmd
			}
		}

		l.viewport, cmd = l.viewport.Update(msg)
		return l, cmd

	case api.ModelsCmd:
		if len(msg.M) > 0 {
			l.models = msg.M
			var m []string
			for _, model := range l.models {
				m = append(m, model.Id)
			}
			slices.Sort(m)
			slices.SortFunc(l.models, func(a, b ai.GroqModelInfo) int {
				return cmp.Compare(a.Id, b.Id)
			})

			config.Config.Model.Options = m
			config.Config.Model.Toggle()
		}

		close(msg.C)
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
		generalTabBtn = library.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render(config.Config.Model.Pretty())
		commandsTabBtn = library.SubtleStyle.Render(config.Config.Effort.Pretty())
	} else {
		generalTabBtn = library.SubtleStyle.Render(config.Config.Model.Pretty())
		commandsTabBtn = library.HighlighedStyle.PaddingLeft(1).PaddingRight(1).Render(config.Config.Effort.Pretty())
	}

	generalTabBtn = zone.Mark("model-toggle", generalTabBtn)
	commandsTabBtn = zone.Mark("effort-toggle", commandsTabBtn)

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, generalTabBtn, " ", commandsTabBtn)

	header := lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		library.BwLabelStyle.MarginRight(1).Render(" /model "),
		tabs,
		"  Compatable models from Groq (console.groq.com/docs/models)",
		library.SubtleStyle.MarginLeft(5).Render("press 'esc' to close"),
		),
	)

	var modelCards []string
	for i, model := range l.models {
		highlight := config.Config.Model.Name == l.models[i].Id
		modelCards = append(modelCards, modelCard(model, highlight))
	}

	l.viewport.SetContent(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			strings.Join(modelCards, "\n\n"),
		),
	)

	return l.viewport.View()
}


func modelCard(modelInfo ai.GroqModelInfo, highlight bool) string {
	header := lipgloss.NewStyle().Render(
		library.BoldStyle.Render(library.SubtleStyle.Render("\u27E1")) +
			" " + modelInfo.Name +
			" (" + modelInfo.Id + ") " +
			library.SubtleStyle.Render(modelInfo.OwnedBy),
	)

	card := header

	if highlight {
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
					library.SubtleStyle.Render(val+" ")+s+"  ",
				)
			}
		}

		capabilities := lipgloss.NewStyle().Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				library.SubtleStyle.Render("Capabilities"),
				lipgloss.JoinHorizontal(lipgloss.Left, labels...),
			),
		)

		card = lipgloss.JoinHorizontal(
			lipgloss.Left,
			library.HighlighedStyle.Height(3).MarginRight(1).Render(library.BoldStyle.Render(" > ")),
			lipgloss.JoinVertical(
				lipgloss.Top,
				header,
				capabilities,
			),
		)
	}		

	return card
}