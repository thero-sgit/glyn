package utils

import (
	"os"
	"path/filepath"
	"strings"
)

type FileEntry struct {
	Name 	 string
	Path     string
	IsDir    bool
}

func ListMatches(token string) ([]FileEntry, error) {
	dir, prefix := filepath.Split(token)

	if dir == "" {
		dir = "."
	}

	destination, err := os.ReadDir(dir)
	if err != nil {
		return []FileEntry{}, err
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

	return entries, nil
}