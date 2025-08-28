package config

import "os"

var (
	KUDOS_SLASH_COMMAND = os.Getenv("KUDOS_SLASH_COMMAND")
	PORT                = os.Getenv("PORT")
	
	// Google Chat OAuth configuration
	GOOGLE_CLIENT_ID     = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_CLIENT_SECRET = os.Getenv("GOOGLE_CLIENT_SECRET")
	GOOGLE_PROJECT_ID    = os.Getenv("GOOGLE_PROJECT_ID")
	GOOGLE_REDIRECT_URI  = os.Getenv("GOOGLE_REDIRECT_URI")
	
	// Google Chat specific configuration
	GOOGLE_CHAT_WEBHOOK_TOKEN = os.Getenv("GOOGLE_CHAT_WEBHOOK_TOKEN")
)