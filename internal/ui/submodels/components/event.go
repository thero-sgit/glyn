package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thero-sgit/glyn/internal/ai/agent"
	"github.com/thero-sgit/glyn/internal/ui/api"
)

type UIEvent struct {
	HistoryIdx int
	spinner    spinner.Model
	e          agent.Event
}

func NewUIEvent(index int, event agent.Event) *UIEvent {
	s := spinner.New()
	s.Spinner = spinner.Pulse

	return &UIEvent{
		HistoryIdx: index,
		spinner: s,
		e: event,
	}
}

func (u *UIEvent) Init() tea.Cmd {
	return nil
}

func (u *UIEvent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case api.TickMsg60:
		var cmd tea.Cmd
		u.spinner, cmd = u.spinner.Update(spinner.TickMsg{})

		return u, cmd

	case agent.Event:
		switch e := msg.(type) {
		case agent.ToolStart:
			u.e = e
		}

		return u, nil
	}

	return u, nil
}

func (u *UIEvent) View() string {
	return fmt.Sprintf("%s %s", u.spinner.View(), u.e.Pretty())
}