package common

var (
	SlashCommands = []Match {
		NewSlashCommand("/help", "/help  "),
		NewSlashCommand("/status", "/status"),
		NewSlashCommand("/model", "/model"),
	}
)