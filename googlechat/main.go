package main

import (
	"fmt"
	"net/http"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"github.com/developertom01/go-kudos/googlechat/config"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/chat/v1"
)

func main() {
	services := services.NewKudosService()
	
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
	
	// OAuth endpoints for Google Chat app installation
	r.GET("/auth/googlechat", handleGoogleChatLogin)
	r.GET("/auth/googlechat/callback", func(c *gin.Context) {
		if database == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
			return
		}
		handleGoogleChatCallback(c, database)
	})
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{"status": "ok", "database": "disconnected"}
		if database != nil {
			status["database"] = "connected"
		}
		c.JSON(200, status)
	})

	// Google Chat webhook endpoint for slash commands
	r.POST("/googlechat/webhook", func(c *gin.Context) {
		if database == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
			return
		}
		
		var event GoogleChatEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Check if this is a message event with slash command
		if event.Type != "MESSAGE" {
			// Return empty response for non-message events
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		// Check if the message contains a slash command
		if event.Message.Text == "" || !isKudosCommand(event.Message.Text) {
			// Return empty response for non-kudos messages
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		response, err := handleGoogleChatCommand(event, services, database)
		if err != nil {
			// Send error message back to chat
			errorResponse := &chat.Message{
				Text: fmt.Sprintf("Error: %s", err.Error()),
			}
			c.JSON(http.StatusOK, errorResponse)
			return
		}

		c.JSON(http.StatusOK, response)
	})

	fmt.Printf("Starting Google Chat server on port %s\n", config.PORT)
	r.Run(config.PORT)
}

// isKudosCommand checks if the message text contains a kudos command
func isKudosCommand(text string) bool {
	if len(text) < 6 {
		return false
	}
	return text == "/kudos" || 
		   (text[:6] == "/kudos" && (len(text) == 6 || text[6] == ' '))
}