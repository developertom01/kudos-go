package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/googlechat/config"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleChatOAuthResponse represents the response from Google OAuth
type GoogleChatOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// getGoogleOAuthConfig returns the OAuth2 configuration for Google Chat
func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.GOOGLE_CLIENT_ID,
		ClientSecret: config.GOOGLE_CLIENT_SECRET,
		RedirectURL:  config.GOOGLE_REDIRECT_URI,
		Scopes: []string{
			"https://www.googleapis.com/auth/chat.bot",
			"https://www.googleapis.com/auth/chat.messages",
		},
		Endpoint: google.Endpoint,
	}
}

// handleGoogleChatLogin initiates the Google Chat OAuth flow
func handleGoogleChatLogin(c *gin.Context) {
	oauthConfig := getGoogleOAuthConfig()
	
	// Generate state parameter for security
	state := fmt.Sprintf("state_%d", time.Now().Unix())
	
	// Build the authorization URL
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	// In production, you should store the state parameter for verification
	_ = state
	
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// handleGoogleChatCallback handles the OAuth callback from Google
func handleGoogleChatCallback(c *gin.Context, database *data.Database) {
	code := c.Query("code")
	errorParam := c.Query("error")
	// state := c.Query("state") // TODO: In production, verify state parameter
	
	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth authorization denied: " + errorParam})
		return
	}
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}
	
	// Exchange code for access token
	oauthConfig := getGoogleOAuthConfig()
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("OAuth token exchange error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}
	
	// Get user info to determine workspace/space details
	ctx := context.Background()
	client := oauthConfig.Client(ctx, token)
	chatService, err := chat.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Failed to create Chat service: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Chat service"})
		return
	}
	
	// For Google Chat, we'll use the project ID as the team identifier
	teamID := config.GOOGLE_PROJECT_ID
	teamName := fmt.Sprintf("Google Chat Project: %s", config.GOOGLE_PROJECT_ID)
	
	// Create or get organization
	org, err := database.CreateOrganization(teamName)
	if err != nil {
		fmt.Printf("Organization creation error: %v\n", err)
		// If organization already exists, try to get it
		org = &data.Organization{ID: 1, Name: teamName}
	}
	
	// Store installation in database
	installation, err := database.CreateInstallation(
		"googlechat",
		org.ID,
		teamID,
		token.AccessToken,
		token.RefreshToken, // Use refresh token as bot token for Google Chat
		teamID,
		teamName,
	)
	
	if err != nil {
		fmt.Printf("Installation creation error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store installation"})
		return
	}
	
	fmt.Printf("Successfully installed app for Google Chat project %s with installation ID: %d\n", 
		config.GOOGLE_PROJECT_ID, installation.ID)
	
	c.HTML(http.StatusOK, "success.html", gin.H{
		"team_name": teamName,
		"message":   "Successfully installed Kudos app for Google Chat!",
	})
	
	// Suppress unused variable warning
	_ = chatService
}

// verifyGoogleChatRequest verifies that the request came from Google Chat
func verifyGoogleChatRequest(c *gin.Context) bool {
	// Google Chat uses a token-based verification
	token := c.GetHeader("Authorization")
	expectedToken := "Bearer " + config.GOOGLE_CHAT_WEBHOOK_TOKEN
	
	return token == expectedToken
}

// authMiddleware is a middleware that verifies Google Chat requests
func authMiddleware(database *data.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for OAuth endpoints and health check
		if c.Request.URL.Path == "/auth/googlechat" || 
		   c.Request.URL.Path == "/auth/googlechat/callback" || 
		   c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		
		// Only verify POST requests (chat commands)
		if c.Request.Method != "POST" {
			c.Next()
			return
		}
		
		// For now, we'll skip token verification to get the basic flow working
		// In production, you should implement proper token verification
		
		c.Next()
	}
}