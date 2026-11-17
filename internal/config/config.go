package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/thero-sgit/glyn/internal/common"
)

var modes = []string{"ask ", "agent"}
var noSlug = regexp.MustCompile(`[^a-z0-9]+`)

type toggle struct {
	Icon    string
	Name    string
	i       int
	Options []string
}

func NewToggle(icon string, options []string) toggle {
	t := toggle{Icon: icon, Options: options}
	t.Name = t.Options[t.i]
	return t
}

func (m *toggle) Toggle() {
	m.i = (m.i + 1) % len(m.Options)
	m.Name = m.Options[m.i]
}

func (m toggle) Pretty() string {
	return m.Icon + " " + m.Name
}

var Config configuration

type configuration struct {
	Version     string
	WorkingDir  string
	DataBaseDir string
	SanitizedWd string
	GroqApiKey  string
	Model       toggle
	Effort      toggle
	Mode        toggle
	MdRenderer  glamour.TermRenderer
	ProjectKey  string
}

func getWd() string {
	if len(os.Args) < 2 {
		log.Fatalf("Please provide a directory path. Usage: xyn <path>")
	}
	path := os.Args[1]

	if strings.Trim(path, " ") == "." {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Could not open directory: %s", err.Error())
		}

		path = wd
	}

	cleanPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Could not open directory: %s", err.Error())
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		log.Fatalf("Could not open directory: %s", err.Error())
	}

	if !info.IsDir() {
		log.Fatalf("%s is not a directory", cleanPath)
	}

	return cleanPath
}

func apiKey() string {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatalf("GROQ_API_KEY is missing. Get your api key from Groq (https://console.groq.com/keys)")
	}

	return apiKey
}

func (c *configuration) SetRenderer(width int) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width-2),
	)
	if err != nil {
		log.Fatal(err)
	}

	c.MdRenderer = *r
}

func projectRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return start
		}

		dir = parent
	}
}

func projectKey(cwd string) (string, error) {
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

func dataBaseDir() (string, error) {
	if d := os.Getenv("GLYN_HOME"); d != "" {
		return d, nil
	}

	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "glyn", "sessions"), nil
	}

	var base string
	switch runtime.GOOS {
	case "windows", "darwin":
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(base, "glyn", "sessions"), nil
}

func Init() {
	Config = configuration{
		Version:    "v1.0.0",
		WorkingDir: getWd(),
		GroqApiKey: apiKey(),
		Model:      NewToggle("\u27E1", []string{""}),
		Effort:     NewToggle("\u26CF", []string{"medium", "high", "low"}),
	}

	base, e := dataBaseDir()
	if e != nil {
		log.Fatal(e)
	}

	if e := os.MkdirAll(base, 0755); e != nil {
		panic(fmt.Sprintf("error creating data:\n\t %s", e.Error()))
	}

	Config.DataBaseDir = base
	Config.SanitizedWd = common.RelativePath(Config.WorkingDir)
	Config.ProjectKey, _ = projectKey(Config.WorkingDir)
	Config.Mode = NewToggle(" ?", modes)
}
