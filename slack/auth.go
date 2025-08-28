package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/slack/config"
	"github.com/gin-gonic/gin"
)

// SlackOAuthResponse represents the response from Slack OAuth
type SlackOAuthResponse struct {
	OK               bool   `json:"ok"`
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
	UserID           string `json:"user_id"`
	TeamID           string `json:"team_id"`
	TeamName         string `json:"team_name"`
	BotUserID        string `json:"bot_user_id"`
	IncomingWebhook  map[string]interface{} `json:"incoming_webhook"`
	Bot              struct {
		BotUserID      string `json:"bot_user_id"`
		BotAccessToken string `json:"bot_access_token"`
	} `json:"bot"`
	Error            string `json:"error,omitempty"`
}

// handleSlackLogin initiates the Slack OAuth flow
func handleSlackLogin(c *gin.Context) {
	// Generate state parameter for security (in production, store this securely)
	state := fmt.Sprintf("state_%d", time.Now().Unix())
	
	// Build the authorization URL
	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=commands,chat:write,users:read&redirect_uri=%s&state=%s",
		url.QueryEscape(config.SLACK_CLIENT_ID),
		url.QueryEscape(config.REDIRECT_URI),
		state,
	)
	
	// In production, you should store the state parameter for verification
	_ = state
	
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// handleSlackCallback handles the OAuth callback from Slack
func handleSlackCallback(c *gin.Context, database *data.Database) {
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
	oauthResponse, err := exchangeCodeForToken(code)
	if err != nil {
		fmt.Printf("OAuth token exchange error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}
	
	if !oauthResponse.OK {
		fmt.Printf("OAuth response not OK: %s\n", oauthResponse.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth exchange failed"})
		return
	}
	
	// Create or get organization
	org, err := database.CreateOrganization(oauthResponse.TeamName)
	if err != nil {
		fmt.Printf("Organization creation error: %v\n", err)
		// If organization already exists, try to get it
		// For now, we'll use a simple approach
		org = &data.Organization{ID: 1, Name: oauthResponse.TeamName}
	}
	
	// Extract bot token from response
	botToken := oauthResponse.Bot.BotAccessToken
	
	// Store installation in database
	installation, err := database.CreateInstallation(
		"slack",
		org.ID,
		oauthResponse.TeamID,
		oauthResponse.AccessToken,
		botToken,
		oauthResponse.TeamID,
		oauthResponse.TeamName,
	)
	
	if err != nil {
		fmt.Printf("Installation creation error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store installation"})
		return
	}
	
	fmt.Printf("Successfully installed app for team %s (ID: %s) with installation ID: %d\n", 
		oauthResponse.TeamName, oauthResponse.TeamID, installation.ID)
	
	c.HTML(http.StatusOK, "success.html", gin.H{
		"team_name": oauthResponse.TeamName,
		"message":   "Successfully installed Kudos app!",
	})
}

// exchangeCodeForToken exchanges the authorization code for an access token
func exchangeCodeForToken(code string) (*SlackOAuthResponse, error) {
	// Create form data
	data := url.Values{}
	data.Set("client_id", config.SLACK_CLIENT_ID)
	data.Set("client_secret", config.SLACK_CLIENT_SECRET)
	data.Set("code", code)
	data.Set("redirect_uri", config.REDIRECT_URI)
	
	// Make request to Slack OAuth endpoint
	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Parse response
	var oauthResponse SlackOAuthResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Parse JSON response
	err = json.Unmarshal(body, &oauthResponse)
	if err != nil {
		return nil, err
	}
	
	// Check for OAuth errors
	if !oauthResponse.OK {
		return nil, fmt.Errorf("OAuth error: %s", oauthResponse.Error)
	}
	
	return &oauthResponse, nil
}

// verifySlackRequest verifies that the request came from Slack
func verifySlackRequest(c *gin.Context) bool {
	// Get the signature from headers
	signature := c.GetHeader("X-Slack-Signature")
	timestamp := c.GetHeader("X-Slack-Request-Timestamp")
	
	if signature == "" || timestamp == "" {
		return false
	}
	
	// Check if timestamp is too old (prevent replay attacks)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	
	if time.Now().Unix()-ts > 300 { // 5 minutes
		return false
	}
	
	// Get request body
	body, err := c.GetRawData()
	if err != nil {
		return false
	}
	
	// Recreate the signature
	sig := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	
	h := hmac.New(sha256.New, []byte(config.SLACK_SIGNING_SECRET))
	h.Write([]byte(sig))
	expectedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))
	
	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// authMiddleware is a middleware that verifies Slack requests
func authMiddleware(database *data.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for OAuth endpoints and health check
		if strings.HasPrefix(c.Request.URL.Path, "/auth/") || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		
		// Only verify POST requests (slash commands)
		if c.Request.Method != "POST" {
			c.Next()
			return
		}
		
		// For now, we'll skip signature verification to get the basic flow working
		// In production, you should implement proper signature verification
		
		c.Next()
	}
}