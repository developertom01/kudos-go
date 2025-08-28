package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Perform the request
	r.ServeHTTP(w, req)
	
	// Assert the response
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestSlackLoginRedirect(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router
	r := gin.New()
	r.GET("/auth/slack", handleSlackLogin)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/auth/slack", nil)
	w := httptest.NewRecorder()
	
	// Perform the request
	r.ServeHTTP(w, req)
	
	// Assert the response is a redirect
	assert.Equal(t, 307, w.Code) // Temporary redirect
	
	// Check that the Location header contains Slack OAuth URL
	location := w.Header().Get("Location")
	assert.Contains(t, location, "slack.com/oauth/v2/authorize")
	assert.Contains(t, location, "client_id=")
	assert.Contains(t, location, "scope=commands,chat:write")
}