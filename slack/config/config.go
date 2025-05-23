package config

import "os"

var (
	KUDOS_SLASH_COMMAND = os.Getenv("KUDOS_SLASH_COMMAND")
	SLACK_API_TOKEN     = os.Getenv("SLACK_API_TOKEN")
	PORT 			= os.Getenv("PORT")
)
