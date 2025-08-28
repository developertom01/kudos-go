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

func TestGoogleChatLoginRedirect(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a test router
	r := gin.New()
	r.GET("/auth/googlechat", handleGoogleChatLogin)
	
	// Create a test request
	req, _ := http.NewRequest("GET", "/auth/googlechat", nil)
	w := httptest.NewRecorder()
	
	// Perform the request
	r.ServeHTTP(w, req)
	
	// Assert the response is a redirect
	assert.Equal(t, 307, w.Code) // Temporary redirect
	
	// Check that the Location header contains Google OAuth URL
	location := w.Header().Get("Location")
	assert.Contains(t, location, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, location, "client_id=")
	assert.Contains(t, location, "scope=")
}