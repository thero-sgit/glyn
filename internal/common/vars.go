package common

var (
	SlashCommands = []Match {
		NewSlashCommand("/help", "/help  "),
		NewSlashCommand("/session", "/session"),
		NewSlashCommand("/model", "/model"),
	}
)