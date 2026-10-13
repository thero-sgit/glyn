package common

var (
	SlashCommands = []Match {
		NewSlashCommand("help", "/help"),
		NewSlashCommand("sessions", "/sessions"),
		NewSlashCommand("model", "/model"),
	}
)