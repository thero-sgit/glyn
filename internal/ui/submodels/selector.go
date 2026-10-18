package submodels

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/common"
	"github.com/thero-sgit/xyn/internal/ui/library"
)

type Selector struct {
	ti       textinput.Model
	tiPrefix string
	Viewport viewport.Model

	//
	Options   []common.Match
	matchFunc func(string) []common.Match
	Cursor    int

	//
	Focused bool
}

func (s *Selector) Init() tea.Cmd {
	return nil
}

func NewSelector(prompt, tiPrefix, subprompt, placeholder string, matchfunc func(string) []common.Match) Selector {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = placeholder
	ti.Prompt = library.BwLabelStyle.Bold(true).Render(" "+prompt+": ") + " " + library.BoldStyle.Foreground(library.AccentColor).Render(subprompt)

	return Selector{
		ti:        ti,
		tiPrefix:  tiPrefix,
		Options:   matchfunc(tiPrefix),
		matchFunc: matchfunc,
	}
}

func (s *Selector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	var cmd tea.Cmd

	switch key.String() {
	case "enter":
		s.Cursor = 0
		s.Options = s.matchFunc(s.tiPrefix)
		s.ti.Reset()

	case "tab":
		s.ti.SetValue(s.Options[s.Cursor].Name())
		s.ti.CursorEnd()
		s.Options = s.matchFunc(s.tiPrefix + s.ti.Value())

	case "up":
		if s.Cursor > 0 {
			s.Cursor--
		}

	case "down":
		if s.Cursor < len(s.Options)-1 {
			s.Cursor++
		}

	case "left", "right":
		s.ti, cmd = s.ti.Update(msg)

	default:
		s.ti, cmd = s.ti.Update(msg)

		s.ti.CursorEnd()
		s.Options = s.matchFunc(s.tiPrefix + s.ti.Value())
		s.Cursor = 0
	}

	return s, cmd
}

func (s *Selector) View() string {
	var rows []string
	for i, option := range s.Options {
		switch {
		case i == s.Cursor:
			rows = append(rows, library.BoldStyle.Foreground(library.AccentColor).Render("> ")+option.Pretty())
		default:
			rows = append(rows, option.Pretty())
		}
	}

	s.Viewport.SetContent(strings.Join(rows, "\n"))

	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, s.ti.View(), s.Viewport.View()))
}
