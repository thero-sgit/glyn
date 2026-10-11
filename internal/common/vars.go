package common

var (
	SlashCommands = []Match {
		NewSlashCommand("/help", "/help  ", nil),
		NewSlashCommand("/status", "/status", nil),
		NewSlashCommand("/model", "/model", nil),
	}
)