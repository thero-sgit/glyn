package persistence

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var noSlug = regexp.MustCompile(`[^a-z0-9]+`)

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

func openSession(base, projectKey, id string) (*os.File, error) {
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

func newSessionId(ts string) (string, error) {
	b := make([]byte, 4)

	if _, err:= rand.Read(b); err != nil {
		return "", err
	}

	return ts + "-" + hex.EncodeToString(b), nil
}

func projectRoot(start string) string {
	dir := start
	for {
		if _, err :=  os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}

		dir = parent
	}
}

func ProjectKey(cwd string) (string, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}

	root := projectRoot(abs)

	slug := noSlug.ReplaceAllString(strings.ToLower(filepath.Base(root)), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "root"
	}
	if len(slug) > 40 {
		slug = slug[:40]
	}

	sum := sha256.Sum256([]byte(root))
	return slug + "-" + hex.EncodeToString(sum[:4]), nil
}