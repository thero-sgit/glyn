package common

import (
	"strings"
)

// Matching
type Match interface {
	Name() string
}

type SlashCommand struct {
	name   string
	pretty string
}

func (s SlashCommand) Name() string {
	return s.name
}

func NewSlashCommand(name, pretty string) SlashCommand {
	return SlashCommand{name: name, pretty: pretty}
}

func ListCmdMatches(token string) []Match {
	var r []Match
	for _, cmd := range SlashCommands {
		if strings.HasPrefix(strings.ToLower(cmd.Name()), strings.ToLower(token)) { r = append(r, cmd)}
	}

	return r
}