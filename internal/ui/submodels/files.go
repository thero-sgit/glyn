package submodels

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/xyn/internal/ui/stylesandcomponents"
)

type FileEntry struct {
	Name 	 string
	Path     string
	IsDir    bool
}

type FileSelector struct {
	ti       textinput.Model
	Viewport viewport.Model
	Prompt   string

	//
	Options   []FileEntry
	Cursor    int

	//
	Focused bool
}

func (f *FileSelector) Init() tea.Cmd {
	return nil
}

func NewFileSelector() FileSelector {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "<filename> to add / rm + space to remove file"

	fib := FileSelector {
		ti:        ti,
		Options:   listMatches(""),
	}

	fib.SetPrompt("@")
	fib.renderViewport()

	return fib
}

func (f *FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}

	var cmd tea.Cmd

	switch key.String() {
	case "enter", "tab":
		f.ti.SetValue(f.Options[f.Cursor].Path)
		f.ti.CursorEnd()
		f.Options = listMatches(f.ti.Value())

	case "up":
		if f.Cursor > 0 {
			f.Cursor--
			
			if f.Cursor < 3 {
				f.Viewport, cmd = f.Viewport.Update(msg)
			}			
		}

	case "down":
		if f.Cursor < len(f.Options)-1 {
			f.Cursor++

			if f.Cursor > 4 {
				f.Viewport, cmd = f.Viewport.Update(msg)
			}			
		}

	default:
		f.ti, cmd = f.ti.Update(msg)
		f.Viewport.GotoTop()
		f.ti.CursorEnd()

		if f.ti.Value() == "rm " {
			f.SetPrompt("rm ")
		}
		
		f.Options = listMatches(f.ti.Value())
		f.Cursor = 0
	}

	f.renderViewport()
	return f, cmd
}

func (f *FileSelector) View() string {
	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, f.ti.View(), f.Viewport.View()))
}

func (f *FileSelector) renderViewport() {
	var rows []string
		for i, option := range f.Options {
			switch {
			case i == f.Cursor:
				rows = append(rows, stylesandcomponents.BoldStyle.Foreground(stylesandcomponents.AccentColor).Render("> ")+option.Name)
			default:
				rows = append(rows, option.Name)
			}
		}

	f.Viewport.SetContent(strings.Join(rows, "\n"))
}

func (f *FileSelector) SetPrompt(p string) {
	label := " Add file: "

	if p == "rm " {
		label = " Remove file: "
		f.ti.Placeholder = "<filename> to remove"
	}

	f.ti.Prompt = stylesandcomponents.BwLabelStyle.Bold(true).Render(label) + lipgloss.NewStyle().Foreground(stylesandcomponents.AccentColor).Bold(true).Render(" "+p)
	f.Prompt = p
	f.ti.Reset()
}


func listMatches(token string) []FileEntry {
	dir, prefix := filepath.Split(token)

	if dir == "" {
		dir = "."
	}

	destination, err := os.ReadDir(dir)
	if err != nil {
		return []FileEntry{}
	}

	var entries []FileEntry
	for _, e := range destination {
		name := e.Name()

		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			path_ := filepath.Join(dir, name)
			if e.IsDir() {
				name = name + "/"
				path_ = path_ + "/"
			}

			entries = append(entries, FileEntry{name, path_, e.IsDir()})
		}		
	}

	return entries
}
