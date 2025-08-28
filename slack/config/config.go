package config

import "os"

var (
	KUDOS_SLASH_COMMAND = os.Getenv("KUDOS_SLASH_COMMAND")
	SLACK_API_TOKEN     = os.Getenv("SLACK_API_TOKEN")
	PORT 			= os.Getenv("PORT")
	
	// OAuth configuration
	SLACK_CLIENT_ID     = os.Getenv("SLACK_CLIENT_ID")
	SLACK_CLIENT_SECRET = os.Getenv("SLACK_CLIENT_SECRET")
	SLACK_SIGNING_SECRET = os.Getenv("SLACK_SIGNING_SECRET")
	REDIRECT_URI        = os.Getenv("REDIRECT_URI")
)
