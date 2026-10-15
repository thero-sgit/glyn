package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
)

var modes = []string{"agent", "ask "}

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
	Version          string
	WorkingDir       string
	SanitizedWd      string
	GroqApiKey       string
	Model 			 toggle
	Effort           toggle
	Mode             toggle
	MdRenderer       glamour.TermRenderer
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

func workingDirSanitized() string {
	path := Config.WorkingDir
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

func Init() {
	Config = configuration {
		Version:    "v1.0.0",
		WorkingDir: getWd(),
		GroqApiKey: apiKey(),
		Model:		NewToggle("\u27E1", []string{""}),
		Effort:     NewToggle("\u26CF", []string{"medium", "high", "low"}),
	}

	Config.SanitizedWd = workingDirSanitized()
	Config.Mode = NewToggle("\u2726", modes)
}