package config

import "os"

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	KUDOS_SLASH_COMMAND = getEnvWithDefault("KUDOS_SLASH_COMMAND", "/kudos")
	PORT                = getEnvWithDefault("PORT", ":8081") // Different port from Slack
	
	// Google Chat OAuth configuration
	GOOGLE_CLIENT_ID     = os.Getenv("GOOGLE_CLIENT_ID")
	GOOGLE_CLIENT_SECRET = os.Getenv("GOOGLE_CLIENT_SECRET")
	GOOGLE_PROJECT_ID    = os.Getenv("GOOGLE_PROJECT_ID")
	GOOGLE_REDIRECT_URI  = getEnvWithDefault("GOOGLE_REDIRECT_URI", "http://localhost:8081/auth/googlechat/callback")
	
	// Google Chat specific configuration
	GOOGLE_CHAT_WEBHOOK_TOKEN = os.Getenv("GOOGLE_CHAT_WEBHOOK_TOKEN")
)