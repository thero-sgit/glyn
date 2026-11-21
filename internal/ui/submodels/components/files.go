package components

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thero-sgit/glyn/internal/ui/library"
)

type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
}

type FileSelector struct {
	ti       textinput.Model
	Viewport viewport.Model
	Prompt   string

	//
	Options      []FileEntry
	Cursor       int
	SelectedFile *FileEntry

	cachedDirs map[string][]FileEntry

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

	fib := FileSelector{
		ti:         ti,
		cachedDirs: make(map[string][]FileEntry),
	}

	fib.SetPrompt("@")
	fib.refreshOptions(true)
	fib.renderViewport()

	return fib
}

func (f *FileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return f, nil
	}
	f.SelectedFile = nil

	var cmd tea.Cmd

	switch key.String() {
	case "enter":
		f.refreshOptions(true)
		if len(f.Options) == 0 {
			break
		}

		selected := f.Options[f.Cursor]
		if selected.IsDir {
			f.ti.SetValue(selected.Path)
			f.ti.CursorEnd()
			f.Viewport.GotoTop()
			f.Cursor = 0
			f.refreshOptions(true)
		} else {
			f.SelectedFile = &selected
			f.ti.Reset()
		}

	case "tab":
		f.refreshOptions(true)
		f.onEnterTab()

	case "up":
		if f.Cursor > 0 {
			f.Cursor--

			if f.Cursor < 5 {
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

	case "left", "right":
		f.ti, cmd = f.ti.Update(msg)
		f.refreshOptions(false)

	default:
		f.ti, cmd = f.ti.Update(msg)
		f.Viewport.GotoTop()
		f.ti.CursorEnd()
		f.Cursor = 0
		f.refreshOptions(false)
	}

	f.SetOptions()
	return f, cmd
}

func (f *FileSelector) View() string {
	return lipgloss.NewStyle().MarginBottom(1).Render(lipgloss.JoinVertical(lipgloss.Top, f.ti.View(), f.Viewport.View()))
}

func (f *FileSelector) SetOptions() {
	f.renderViewport()
}

func (f *FileSelector) onEnterTab() {
	if len(f.Options) > 0 {
		f.ti.SetValue(f.Options[f.Cursor].Path)
		f.Cursor = 0
		f.Viewport.GotoTop()
		f.ti.CursorEnd()
		f.refreshOptions(true)
	}
}

func (f *FileSelector) renderViewport() {
	var rows []string
	for i, option := range f.Options {
		switch {
		case i == f.Cursor:
			rows = append(rows, library.BoldStyle.Foreground(library.AccentColor).Render("> ")+option.Name)
		default:
			rows = append(rows, option.Name)
		}
	}

	f.Viewport.SetContent(strings.Join(rows, "\n"))
}

func (f *FileSelector) refreshOptions(readDirectory bool) {
	dir, prefix := filepath.Split(f.ti.Value())
	if dir == "" {
		dir = "."
	}
	dir = filepath.Clean(dir)

	if readDirectory {
		destination, err := os.ReadDir(dir)
		if err != nil {
			delete(f.cachedDirs, dir)
			f.Options = nil
			f.Cursor = 0
			return
		}

		entries := make([]FileEntry, 0, len(destination))
		for _, entry := range destination {
			name := entry.Name()
			path := filepath.Join(dir, name)
			if entry.IsDir() {
				name += string(os.PathSeparator)
				path += string(os.PathSeparator)
			}
			entries = append(entries, FileEntry{
				Name:  name,
				Path:  path,
				IsDir: entry.IsDir(),
			})
		}
		f.cachedDirs[dir] = entries
	}

	f.Options = f.Options[:0]
	prefix = strings.ToLower(prefix)
	for _, entry := range f.cachedDirs[dir] {
		if strings.HasPrefix(strings.ToLower(strings.TrimSuffix(entry.Name, string(os.PathSeparator))), prefix) {
			f.Options = append(f.Options, entry)
		}
	}
	if f.Cursor >= len(f.Options) {
		f.Cursor = len(f.Options) - 1
	}
	if f.Cursor < 0 {
		f.Cursor = 0
	}
}

func (f *FileSelector) SetPrompt(p string) {
	label := " Add file: "

	f.ti.Prompt = library.BwLabelStyle.Bold(true).Render(label) + lipgloss.NewStyle().Foreground(library.AccentColor).Bold(true).Render(" "+p)
	f.Prompt = p
	f.ti.Reset()
}
