package persistence

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type ChatSession struct {
	Id          string
	Name        string
	Model       string
	LastUpdated string
}

func NewChatSession(name, id, model, lastUpdated string) *ChatSession {
	return &ChatSession {
		Id: id,
		Name: name,
		Model: model,
		LastUpdated: lastUpdated,
	}
}

// func NewChatSessionFromId(key, id string) (*ChatSession, error) {
// 	file, _ := os.Open(filepath.Join(config.Config.DataBaseDir, key, id + ".jsonl"))
// 	defer file.Close()

// 	reader := bufio.NewReaderSize(file, 1<<20)
// 	reader.ReadBytes('\n')

// 	line, err := reader.ReadBytes('\n')
// 	if err != nil && err != io.EOF {
// 		return &ChatSession{}, err
// 	}

// 	var title TitleEntry
// 	if err := json.Unmarshal(line, &title); err != nil {
// 		return &ChatSession{}, err
// 	}

// 	return NewChatSession(title.Value), nil
// }

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
			style.Render(" · " + c.LastUpdated),
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