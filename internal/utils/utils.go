package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func listMatches(token string) ([]os.DirEntry, error) {
	dir, prefix := filepath.Split(token)

	if dir == "" {
		dir = "."
	}

	destination, err := os.ReadDir(dir)
	if err != nil {
		return []os.DirEntry{}, err
	}

	var entries []os.DirEntry
	for _, e := range destination {
		if strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(prefix)) {
			entries = append(entries, e)
		}		
	}

	return entries, nil
}