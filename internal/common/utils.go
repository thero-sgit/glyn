package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Matching
type Match interface {
	Name() 	 string
	Pretty() string
}

type SlashCommand struct {
	name   string
	pretty string
}

func (s SlashCommand) Name() string {
	return s.name
}

func (s SlashCommand) Pretty() string {
	return s.pretty
}

func NewSlashCommand(name, pretty string) SlashCommand {
	return SlashCommand{name: name, pretty: pretty}
}

func ListCmdMatches(token string) []Match {
	var r []Match
	for _, cmd := range SlashCommands {
		if strings.HasPrefix(strings.ToLower("/" + cmd.Name()), strings.ToLower(token)) { r = append(r, cmd)}
	}

	return r
}

func SafePath(root, path string) (string, error) {
	abs := filepath.Join(root, path)
	if filepath.IsAbs(abs) {
		abs = filepath.Clean(abs)
	}

	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return  "", fmt.Errorf("path %q is outside the workspace", path)
	}

	return abs, nil
}

func BinaryCheck(file *os.File) (*os.File, error) {
	head := make([]byte, 8192)
	n, _ := file.Read(head)
	if bytes.IndexByte(head[:n], 0) >= 0 {
		return nil, errors.New("binary file, not shown")
	}
	file.Seek(0, io.SeekStart)

	return file, nil
}


func Truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    cut := s[:max]
    if i := strings.LastIndexByte(cut, '\n'); i > 0 {
        cut = cut[:i]
    }
    
    cut = strings.ToValidUTF8(cut, "")
    return cut + fmt.Sprintf(
        "\n\n[output truncated: showed %d of %d bytes. Narrow your search or read a smaller range.]",
        len(cut), len(s),
    )
}

func RelativePath(path string) string {
	splitPath := strings.Split(path, "/")

	if len(splitPath) < 2 {
		return path
	}

	return strings.Join(
		[]string{
			"...",
			splitPath[len(splitPath)-2],
			splitPath[len(splitPath)-1],
		},
		"/",
	)
}