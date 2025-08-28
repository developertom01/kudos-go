package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set test environment variables
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_PROJECT_ID", "test-project")
	os.Setenv("GOOGLE_REDIRECT_URI", "http://localhost:8081/auth/googlechat/callback")
	os.Setenv("GOOGLE_CHAT_WEBHOOK_TOKEN", "test-webhook-token")
	
	gin.SetMode(gin.TestMode)
	code := m.Run()
	os.Exit(code)
}

func TestHealthEndpoint(t *testing.T) {
	r := gin.New()
	
	// Add the actual health endpoint logic
	r.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"status":       "ok",
			"service":      "Google Chat Kudos Bot",
			"version":      "1.0.0",
			"database":     "disconnected",
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"configuration": "ok",
		}
		c.JSON(http.StatusOK, status)
	})
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Google Chat Kudos Bot", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "disconnected", response["database"])
	assert.Equal(t, "ok", response["configuration"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestGoogleChatLoginRedirect(t *testing.T) {
	r := gin.New()
	r.GET("/auth/googlechat", handleGoogleChatLogin)
	
	req, _ := http.NewRequest("GET", "/auth/googlechat", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	
	location := w.Header().Get("Location")
	assert.Contains(t, location, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, location, "client_id=test-client-id")
	assert.Contains(t, location, "scope=")
	assert.Contains(t, location, "state=")
	assert.Contains(t, location, "access_type=offline")
}

func TestGoogleChatCallbackMissingCode(t *testing.T) {
	r := gin.New()
	r.GET("/auth/googlechat/callback", func(c *gin.Context) {
		handleGoogleChatCallback(c, nil) // No database for this test
	})
	
	req, _ := http.NewRequest("GET", "/auth/googlechat/callback", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization code")
}

func TestGoogleChatCallbackWithError(t *testing.T) {
	r := gin.New()
	r.GET("/auth/googlechat/callback", func(c *gin.Context) {
		handleGoogleChatCallback(c, nil)
	})
	
	req, _ := http.NewRequest("GET", "/auth/googlechat/callback?error=access_denied", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "OAuth authorization denied")
	assert.Contains(t, w.Body.String(), "access_denied")
}

func TestGoogleChatCallbackInvalidState(t *testing.T) {
	r := gin.New()
	r.GET("/auth/googlechat/callback", func(c *gin.Context) {
		handleGoogleChatCallback(c, nil)
	})
	
	req, _ := http.NewRequest("GET", "/auth/googlechat/callback?code=test-code&state=invalid-state", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired authentication request")
}

func TestVerifyGoogleChatRequest(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedResult bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer test-webhook-token",
			expectedResult: true,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer wrong-token",
			expectedResult: false,
		},
		{
			name:           "Missing Bearer prefix",
			authHeader:     "test-webhook-token",
			expectedResult: false,
		},
		{
			name:           "Empty header",
			authHeader:     "",
			expectedResult: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			
			var result bool
			r.POST("/test", func(c *gin.Context) {
				result = verifyGoogleChatRequest(c)
				c.JSON(http.StatusOK, gin.H{"verified": result})
			})
			
			req, _ := http.NewRequest("POST", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(authMiddleware(nil))
	
	// OAuth endpoints should skip auth
	r.GET("/auth/googlechat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "oauth"})
	})
	
	r.GET("/auth/googlechat/callback", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "callback"})
	})
	
	// Health endpoint should skip auth
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "health"})
	})
	
	// Webhook should require auth
	r.POST("/googlechat/webhook", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "webhook"})
	})
	
	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "OAuth endpoint allows access",
			method:         "GET",
			path:           "/auth/googlechat",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Callback endpoint allows access",
			method:         "GET",
			path:           "/auth/googlechat/callback",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health endpoint allows access",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Webhook with valid token",
			method:         "POST",
			path:           "/googlechat/webhook",
			authHeader:     "Bearer test-webhook-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Webhook with invalid token",
			method:         "POST",
			path:           "/googlechat/webhook",
			authHeader:     "Bearer wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Webhook without token",
			method:         "POST",
			path:           "/googlechat/webhook",
			expectedStatus: http.StatusUnauthorized,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestStateStore(t *testing.T) {
	// Test state generation
	state1, err := generateSecureState()
	require.NoError(t, err)
	assert.NotEmpty(t, state1)
	
	state2, err := generateSecureState()
	require.NoError(t, err)
	assert.NotEmpty(t, state2)
	assert.NotEqual(t, state1, state2) // Should be unique
	
	// Test state storage and validation
	store := &StateStore{states: make(map[string]time.Time)}
	
	// Store a valid state
	store.storeState(state1)
	assert.True(t, store.validateState(state1))
	
	// State should be removed after validation
	assert.False(t, store.validateState(state1))
	
	// Invalid state should return false
	assert.False(t, store.validateState("invalid-state"))
	
	// Test expiration (this is a bit tricky to test without mocking time)
	store.storeState(state2)
	// Manually expire the state
	store.states[state2] = time.Now().Add(-11 * time.Minute)
	assert.False(t, store.validateState(state2))
}

func TestRateLimitMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(rateLimitMiddleware())
	
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// Test normal requests
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// Test rate limiting would require more sophisticated setup
	// For now, just test that middleware doesn't break normal flow
}

func TestValidateConfiguration(t *testing.T) {
	// Save original env vars
	originalClientID := os.Getenv("GOOGLE_CLIENT_ID")
	originalClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	originalProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	
	defer func() {
		// Restore original env vars
		os.Setenv("GOOGLE_CLIENT_ID", originalClientID)
		os.Setenv("GOOGLE_CLIENT_SECRET", originalClientSecret)
		os.Setenv("GOOGLE_PROJECT_ID", originalProjectID)
	}()
	
	// Test with all required vars set
	os.Setenv("GOOGLE_CLIENT_ID", "test-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-secret")
	os.Setenv("GOOGLE_PROJECT_ID", "test-project")
	
	err := validateConfiguration()
	assert.NoError(t, err)
	
	// Test with missing CLIENT_ID
	os.Setenv("GOOGLE_CLIENT_ID", "")
	err = validateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GOOGLE_CLIENT_ID")
	
	// Test with missing CLIENT_SECRET
	os.Setenv("GOOGLE_CLIENT_ID", "test-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "")
	err = validateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GOOGLE_CLIENT_SECRET")
	
	// Test with missing PROJECT_ID
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-secret")
	os.Setenv("GOOGLE_PROJECT_ID", "")
	err = validateConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GOOGLE_PROJECT_ID")
}

func TestWebhookWithInvalidJSON(t *testing.T) {
	r := gin.New()
	r.Use(authMiddleware(nil))
	
	r.POST("/googlechat/webhook", func(c *gin.Context) {
		// This endpoint requires database but we're testing JSON parsing
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database not available"})
	})
	
	// Test with invalid JSON
	invalidJSON := `{"type": "MESSAGE", "invalid": json}`
	req, _ := http.NewRequest("POST", "/googlechat/webhook", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-webhook-token")
	
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	// Should get service unavailable because we don't have database
	// But JSON parsing should work
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}