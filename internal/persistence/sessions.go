package persistence

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/config"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type ChatSession struct {
	Name        string
	changes     [2]string
	tokenUsage  string
	lastUpdated string
	file        *os.File

	//
	id  string
	key string
}

func NewChatSession(name string) *ChatSession {
	return &ChatSession {
		Name: name,
		tokenUsage: "12,032",
		lastUpdated: "2 hours ago",
	}
}

func NewChatSessionFromId(key, id string) (*ChatSession, error) {
	file, _ := os.Open(filepath.Join(config.Config.DataBaseDir, key, id + ".jsonl"))
	defer file.Close()

	reader := bufio.NewReaderSize(file, 1<<20)
	reader.ReadBytes('\n')

	line, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return &ChatSession{}, err
	}

	var title TitleEntry
	if err := json.Unmarshal(line, &title); err != nil {
		return &ChatSession{}, err
	}

	return NewChatSession(title.Value), nil
}

func (c ChatSession) Pretty(width int, highlight, current bool) string {
	header := "\u25CB  " + c.Name
	if current {
		header = "\u25CF  " + c.Name
	}

	style := lipgloss.NewStyle()
	if highlight {
		style = library.HighlighedStyle
	}

	inf := style.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"   ",
			style.Render("+"+c.changes[0] + " -"+c.changes[1]),
			style.Render(" · " + c.lastUpdated),
			style.Render(" · " + c.tokenUsage + " tokens"),
		),
	)

	return style.Width(width).MarginBottom(1).PaddingLeft(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			header,
			inf,
		),
	)
}