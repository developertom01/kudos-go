package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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

// StateStore manages OAuth state parameters for security
type StateStore struct {
	states map[string]time.Time
}

var stateStore = &StateStore{
	states: make(map[string]time.Time),
}

// generateSecureState generates a cryptographically secure state parameter
func generateSecureState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// storeState stores a state parameter with expiration
func (s *StateStore) storeState(state string) {
	s.states[state] = time.Now().Add(10 * time.Minute) // 10 minute expiration
}

// validateState validates and removes an expired state parameter
func (s *StateStore) validateState(state string) bool {
	if expiry, exists := s.states[state]; exists {
		delete(s.states, state)
		return time.Now().Before(expiry)
	}
	return false
}
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
	
	// Generate secure state parameter for CSRF protection
	state, err := generateSecureState()
	if err != nil {
		fmt.Printf("Failed to generate state: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate authentication"})
		return
	}
	
	// Store state for later verification
	stateStore.storeState(state)
	
	// Build the authorization URL with state parameter
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// handleGoogleChatCallback handles the OAuth callback from Google
func handleGoogleChatCallback(c *gin.Context, database *data.Database) {
	code := c.Query("code")
	errorParam := c.Query("error")
	state := c.Query("state")
	
	if errorParam != "" {
		fmt.Printf("OAuth authorization error: %s\n", errorParam)
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth authorization denied: " + errorParam})
		return
	}
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}
	
	// Verify state parameter for CSRF protection
	if state == "" || !stateStore.validateState(state) {
		fmt.Printf("Invalid or expired state parameter: %s\n", state)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired authentication request"})
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
	
	// Check if webhook token is configured
	if config.GOOGLE_CHAT_WEBHOOK_TOKEN == "" {
		fmt.Printf("Warning: GOOGLE_CHAT_WEBHOOK_TOKEN not configured\n")
		return false
	}
	
	expectedToken := "Bearer " + config.GOOGLE_CHAT_WEBHOOK_TOKEN
	
	if token != expectedToken {
		fmt.Printf("Invalid webhook token. Expected: %s, Got: %s\n", expectedToken, token)
		return false
	}
	
	return true
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
		
		// Only verify POST requests to webhook endpoints
		if c.Request.Method == "POST" && c.Request.URL.Path == "/googlechat/webhook" {
			if !verifyGoogleChatRequest(c) {
				fmt.Printf("Authentication failed for webhook request from %s\n", c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}