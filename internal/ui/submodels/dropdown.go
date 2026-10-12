package submodels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

type Dropdown struct {
	options   []string
	selected int
	cursor   int
}


func (d Dropdown) Init() tea.Cmd {
	return nil
}

func NewDropdown(options []string) tea.Model {
	return Dropdown{
		options: options,
	}
}

func (d Dropdown) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}

	switch key.String() {
	case "enter", " ":
		d.selected = d.cursor

	case "up":
		if d.cursor > 0 {
			d.cursor--
		}

	case "down":
		if d.cursor < len(d.options)-1 {
			d.cursor++
		}
	}

	return d, nil
}

func (d Dropdown) View() string {
	arrow := ">"

	header := stylesandcomponents.HighlighedStyle.Bold(true).Render(fmt.Sprintf(" Header %s ", arrow))

	var rows []string
	for i, option := range d.options {
		switch {
		case i == d.cursor:
			rows = append(rows, stylesandcomponents.HighlighedStyle.Render(fmt.Sprintf(" %s ", option)))
		default:
			rows = append(rows, " "+option)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, strings.Join(rows, "\n"))
}
