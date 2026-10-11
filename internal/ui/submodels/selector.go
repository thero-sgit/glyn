package submodels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/common"
)

type Selector struct {
	ti 			textinput.Model

	//
	Options     []common.Match
	matchFunc   func(string) []common.Match
	Selected    int
	cursor      int

	//
	Focused     bool
}


func (s Selector) Init() tea.Cmd {
	return nil
}

func NewSelector(prompt, placeholder string, options []common.Match, matchfunc func(string) []common.Match) Selector {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.Prompt = common.BwLabelStyle.Bold(true).Render(" " + prompt + " ") 

	return Selector{
		ti:        ti,
		Options:   options,
		matchFunc: matchfunc,
	}
}

func (s Selector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	var cmd tea.Cmd

	switch key.String() {
	case "enter", " ":
		s.Selected = s.cursor

	case "up":
		if s.cursor > 0 {
			s.cursor--
		}

	case "down":
		if s.cursor < len(s.Options)-1 {
			s.cursor++
		}

	default:
		s.ti, cmd = s.ti.Update(msg)
		s.Options = s.matchFunc(s.ti.Value())
	}

	return s, cmd
}

func (s Selector) View() string {

	var rows []string
	for i, option := range s.Options {
		switch {
		case i == s.cursor:
			rows = append(rows, common.HighlighedStyle.Render(fmt.Sprintf(" %s ", option.Name())))
		default:
			rows = append(rows, option.Name())
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, s.ti.View(), strings.Join(rows, "\n"))
}
