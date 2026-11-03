package persistence

import (
	"os"
	"path/filepath"
)

func dataBaseDir() (string, error) {
	if d := os.Getenv("GLYN_HOME"); d != "" {
		return d, nil
	}

	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "glyn"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "glyn"), nil
}

func openSession(projectKey, id string) (*os.File, error) {
	base, err := dataBaseDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(base, "sessions", projectKey)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return os.OpenFile(
		filepath.Join(dir, id+".jsonl"),
		os.O_CREATE | os.O_APPEND | os.O_WRONLY,
		0o600,
	)
}