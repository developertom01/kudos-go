package main

import (
	"fmt"
	"net/http"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"github.com/developertom01/go-kudos/slack/config"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func main() {
	services := services.NewKudosService()
	slackApi := slack.New(config.SLACK_API_TOKEN)
	
	// Try to connect to database, but don't panic if it fails
	database, err := data.NewDatabase("")
	if err != nil {
		fmt.Printf("Warning: Database connection failed: %v\n", err)
		fmt.Println("Running in demo mode without database functionality")
		database = nil
	}

	r := gin.Default()
	
	// Load HTML templates
	r.LoadHTMLGlob("templates/*")
	
	// Add authentication middleware for non-auth routes (skip if no database)
	if database != nil {
		r.Use(authMiddleware(database))
	}
	
	// OAuth endpoints for Slack app installation
	r.GET("/auth/slack", handleSlackLogin)
	r.GET("/auth/slack/callback", func(c *gin.Context) {
		if database == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
			return
		}
		handleSlackCallback(c, database)
	})
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{"status": "ok", "database": "disconnected"}
		if database != nil {
			status["database"] = "connected"
		}
		c.JSON(200, status)
	})

	// Slash command endpoint
	r.POST(config.KUDOS_SLASH_COMMAND, func(c *gin.Context) {
		if database == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
			return
		}
		
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

	fmt.Printf("Starting server on port %s\n", config.PORT)
	r.Run(config.PORT)

}
