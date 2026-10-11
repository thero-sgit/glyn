package submodels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/utils"
)

type Dropdown struct {
	options  []string
	selected int
	cursor   int
	Open     bool
}


func (d *Dropdown) Init() tea.Cmd {
	return nil
}

func NewDropdown(options []string) Dropdown {
	return Dropdown{
		options: options,
	}
}

func (d *Dropdown) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}

	switch key.String() {
	case "enter", " ":
		if d.Open {
			d.selected = d.cursor
			d.Open = false
		} else {
			d.cursor = d.selected
			d.Open = true
		}

	case "esc":
		d.Open = false

	case "up":
		if d.Open && d.cursor > 0 {
			d.cursor--
		}

	case "down":
		if d.Open && d.cursor < len(d.options)-1 {
			d.cursor++
		}
	}

	return d, nil
}

func (d *Dropdown) View() string {
	arrow := "<"
	if d.Open {
		arrow = ">"
	}

	header := utils.HighlighedStyle.Bold(true).Render(fmt.Sprintf(" Header %s ", arrow))
	if !d.Open {
		return header
	}

	var rows []string
	for i, option := range d.options {
		switch {
		case i == d.cursor:
			rows = append(rows, utils.HighlighedStyle.Render(fmt.Sprintf(" %s ", option)))
		default:
			rows = append(rows, " "+option)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, strings.Join(rows, "\n"))
}
