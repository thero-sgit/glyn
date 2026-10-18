package submodels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/xyn/internal/ui/library"
)

type Dropdown struct {
	id       string
	options  []string
	cursor   int
	Selected string
	Focused  bool
	open     bool
}

func (d *Dropdown) Init() tea.Cmd {
	return nil
}

func NewDropdown(id, selected string, options []string) Dropdown {
	return Dropdown{
		id: id,
		Selected: selected,
		options: options,
	}
}

func (d *Dropdown) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if zone.Get(d.id+"-dropdown-menu").InBounds(msg) && d.Focused {
			d.open = !d.open
			return d, nil
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			d.Selected = d.options[d.cursor]
			d.open = false
			return d, nil

		case "up":
			if d.cursor > 0 && d.Focused {
				d.cursor--
			}

		case "down":
			if d.cursor < len(d.options)-1 && d.Focused {
				d.cursor++
			}
		}
	}

	return d, nil
}

func (d *Dropdown) View() string {
	arrow := "\u27A4"
	if d.open {
		arrow = "I"
	}

	header := zone.Mark(d.id+"-dropdown-menu", library.HighlighedStyle.Render(fmt.Sprintf(" %s %s ", d.Selected, library.BoldStyle.Render(arrow+" "))))

	if !d.open {
		return header
	}

	var rows []string
	for i, option := range d.options {
		switch {
		case i == d.cursor:
			rows = append(rows, library.BoldStyle.Foreground(library.AccentColor).Render("> ")+option)
		default:
			rows = append(rows, " "+option)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, strings.Join(rows, "\n"))
}
