package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"github.com/developertom01/go-kudos/googlechat/config"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/chat/v1"
)

// validateConfiguration checks if required environment variables are set
func validateConfiguration() error {
	required := map[string]string{
		"GOOGLE_CLIENT_ID":     config.GOOGLE_CLIENT_ID,
		"GOOGLE_CLIENT_SECRET": config.GOOGLE_CLIENT_SECRET,
		"GOOGLE_PROJECT_ID":    config.GOOGLE_PROJECT_ID,
	}
	
	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}
	
	return nil
}

// rateLimitMiddleware provides basic rate limiting
func rateLimitMiddleware() gin.HandlerFunc {
	// Simple rate limiting - in production, use Redis or more sophisticated solution
	requests := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old requests (older than 1 minute)
		if times, exists := requests[clientIP]; exists {
			var recent []time.Time
			for _, t := range times {
				if now.Sub(t) < time.Minute {
					recent = append(recent, t)
				}
			}
			requests[clientIP] = recent
		}
		
		// Check rate limit (max 30 requests per minute)
		if len(requests[clientIP]) >= 30 {
			log.Printf("Rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		
		// Record this request
		requests[clientIP] = append(requests[clientIP], now)
		c.Next()
	}
}
func main() {
	// Validate configuration first
	if err := validateConfiguration(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	log.Printf("Starting Google Chat Kudos Bot v1.0")
	log.Printf("Port: %s", config.PORT)
	log.Printf("Project ID: %s", config.GOOGLE_PROJECT_ID)
	
	services := services.NewKudosService()
	
	// Try to connect to database, but don't panic if it fails
	database, err := data.NewDatabase("")
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running in demo mode without database functionality")
		database = nil
	} else {
		log.Println("Database connection established")
	}

	// Set Gin mode based on environment
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.Default()
	
	// Add rate limiting middleware
	r.Use(rateLimitMiddleware())
	
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
	
	// Health check endpoint with detailed status
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"status":    "ok",
			"service":   "Google Chat Kudos Bot",
			"version":   "1.0.0",
			"database":  "disconnected",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		
		if database != nil {
			status["database"] = "connected"
		}
		
		// Check required configuration
		configStatus := "ok"
		if config.GOOGLE_CLIENT_ID == "" || config.GOOGLE_CLIENT_SECRET == "" {
			configStatus = "incomplete"
		}
		status["configuration"] = configStatus
		
		httpStatus := http.StatusOK
		if database == nil || configStatus != "ok" {
			httpStatus = http.StatusServiceUnavailable
			status["status"] = "degraded"
		}
		
		c.JSON(httpStatus, status)
	})

	// Google Chat webhook endpoint for slash commands
	r.POST("/googlechat/webhook", func(c *gin.Context) {
		if database == nil {
			log.Printf("Webhook request received but database not available")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
			return
		}
		
		var event GoogleChatEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			log.Printf("Invalid webhook request format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		log.Printf("Received Google Chat event: type=%s, space=%s", event.Type, event.Space.Name)

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

		log.Printf("Processing kudos command: %s", event.Message.Text)

		response, err := handleGoogleChatCommand(event, services, database)
		if err != nil {
			log.Printf("Error processing kudos command: %v", err)
			// Send error message back to chat
			errorResponse := &chat.Message{
				Text: fmt.Sprintf("❌ Error: %s", err.Error()),
			}
			c.JSON(http.StatusOK, errorResponse)
			return
		}

		log.Printf("Kudos command processed successfully")
		c.JSON(http.StatusOK, response)
	})

	log.Printf("Starting Google Chat server on port %s", config.PORT)
	if err := r.Run(config.PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// isKudosCommand checks if the message text contains a kudos command
func isKudosCommand(text string) bool {
	if len(text) < 6 {
		return false
	}
	return text == "/kudos" || 
		   (text[:6] == "/kudos" && (len(text) == 6 || text[6] == ' '))
}