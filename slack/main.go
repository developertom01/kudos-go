package main

import (
	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"github.com/developertom01/go-kudos/slack/config"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func main() {
	services := services.NewKudosService()
	slackApi := slack.New(config.SLACK_API_TOKEN)
	database, err := data.NewDatabase("")

	if err != nil {
		panic(err)
	}

	r := gin.Default()
	
	// Load HTML templates
	r.LoadHTMLGlob("templates/*")
	
	// Add authentication middleware for non-auth routes
	r.Use(authMiddleware(database))
	
	// OAuth endpoints for Slack app installation
	r.GET("/auth/slack", handleSlackLogin)
	r.GET("/auth/slack/callback", func(c *gin.Context) {
		handleSlackCallback(c, database)
	})
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Slash command endpoint
	r.POST(config.KUDOS_SLASH_COMMAND, func(c *gin.Context) {
		slashCommand, err := slack.SlashCommandParse(c.Request)

		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request",
			})
			return
		}

		err = handleSlashCommand(slashCommand, services, slackApi, database)

		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request",
			})
			return
		}
		c.JSON(200, nil)
	})

	r.Run(config.PORT)

}
