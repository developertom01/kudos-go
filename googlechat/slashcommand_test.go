package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/chat/v1"
)

func TestParseCommandText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *Kudos
		shouldError bool
	}{
		{
			name:  "Valid Google Chat @mention format",
			input: "<users/123456789> great work on the project",
			expected: &Kudos{
				Command:     KudosCommand,
				UserID:      "123456789",
				Description: "great work on the project",
			},
			shouldError: false,
		},
		{
			name:  "Valid legacy @username format",
			input: "@john awesome debugging skills",
			expected: &Kudos{
				Command:     KudosCommand,
				Username:    "john",
				Description: "awesome debugging skills",
			},
			shouldError: false,
		},
		{
			name:  "Multi-word description",
			input: "@jane thank you for helping with the complex database optimization task",
			expected: &Kudos{
				Command:     KudosCommand,
				Username:    "jane",
				Description: "thank you for helping with the complex database optimization task",
			},
			shouldError: false,
		},
		{
			name:        "Missing description",
			input:       "@john",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "Missing username",
			input:       "great work",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "No @ prefix",
			input:       "john great work",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "Empty text",
			input:       "",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "Single word (insufficient args)",
			input:       "test",
			expected:    nil,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "command format")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Command, result.Command)
				assert.Equal(t, tt.expected.UserID, result.UserID)
				assert.Equal(t, tt.expected.Username, result.Username)
				assert.Equal(t, tt.expected.Description, result.Description)
			}
		})
	}
}

func TestParseCommandTextGoogleChatMentionFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected UserID
	}{
		{
			name:     "Standard Google Chat user ID",
			input:    "<users/123456789> great work",
			expected: "123456789",
		},
		{
			name:     "Google Chat user ID with different format",
			input:    "<users/987654321> excellent debugging",
			expected: "987654321",
		},
		{
			name:     "Long user ID",
			input:    "<users/123456789012345678> outstanding work",
			expected: "123456789012345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.UserID)
			assert.Empty(t, result.Username) // Should not set username when UserID is set
		})
	}
}

func TestParseCommandTextComplexScenarios(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedDesc string
		shouldError bool
	}{
		{
			name:        "Long description with punctuation",
			input:       "@alice Thank you for the excellent code review! Your feedback was spot-on.",
			expectedDesc: "Thank you for the excellent code review! Your feedback was spot-on.",
			shouldError: false,
		},
		{
			name:        "Description with emojis",
			input:       "<users/123456789> Amazing work on the deployment! 🚀🎉",
			expectedDesc: "Amazing work on the deployment! 🚀🎉",
			shouldError: false,
		},
		{
			name:        "Multiple spaces in description normalized",
			input:       "@bob    Excellent    debugging    skills",
			expectedDesc: "Excellent debugging skills",
			shouldError: false,
		},
		{
			name:        "Leading slash removed correctly",
			input:       "@charlie Great work on the feature",
			expectedDesc: "Great work on the feature",
			shouldError: false,
		},
		{
			name:        "Special characters in description",
			input:       "@dave Thanks for fixing the $variable issue & the #hashtag problem!",
			expectedDesc: "Thanks for fixing the $variable issue & the #hashtag problem!",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDesc, result.Description)
			}
		})
	}
}

func TestParseCommandTextEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Invalid Google Chat mention format - missing closing bracket",
			input:       "<users/123456789 great work",
			shouldError: true,
			errorMsg:    "user must be mentioned",
		},
		{
			name:        "Invalid Google Chat mention format - wrong prefix",
			input:       "<user/123456789> great work",
			shouldError: true,
			errorMsg:    "user must be mentioned",
		},
		{
			name:        "Invalid Google Chat mention format - empty user ID",
			input:       "<users/> great work",
			shouldError: true,
			errorMsg:    "user must be mentioned",
		},
		{
			name:        "Username without @ prefix",
			input:       "alice great work",
			shouldError: true,
			errorMsg:    "user must be mentioned",
		},
		{
			name:        "Just @ symbol",
			input:       "@ great work",
			shouldError: false, // This should parse as username=""
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)
			
			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestIsKudosCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Exact match",
			input:    "/kudos",
			expected: true,
		},
		{
			name:     "Command with arguments",
			input:    "/kudos @user great work",
			expected: true,
		},
		{
			name:     "Command with space",
			input:    "/kudos ",
			expected: true,
		},
		{
			name:     "Different command",
			input:    "/help",
			expected: false,
		},
		{
			name:     "Not a command",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Partial match but not command",
			input:    "kudos to you",
			expected: false,
		},
		{
			name:     "Command prefix but different command",
			input:    "/kudoss @user",
			expected: false,
		},
		{
			name:     "Too short input",
			input:    "/kud",
			expected: false,
		},
		{
			name:     "No slash prefix",
			input:    "kudos @user great",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKudosCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoogleChatEventStructure(t *testing.T) {
	// Test that GoogleChatEvent can be used properly
	event := GoogleChatEvent{
		Type: "MESSAGE",
		Message: struct {
			Name         string `json:"name"`
			Sender       struct {
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
				Type        string `json:"type"`
			} `json:"sender"`
			Text         string `json:"text"`
			ArgumentText string `json:"argumentText"`
			Space        struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"space"`
			Thread struct {
				Name string `json:"name"`
			} `json:"thread"`
		}{
			ArgumentText: "@testuser great work!",
			Sender: struct {
				Name        string `json:"name"`
				DisplayName string `json:"displayName"`
				Type        string `json:"type"`
			}{
				Name:        "users/123456789",
				DisplayName: "Test User",
				Type:        "HUMAN",
			},
		},
		Space: struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}{
			Name: "spaces/test-space",
			Type: "ROOM",
		},
	}
	
	assert.Equal(t, "MESSAGE", event.Type)
	assert.Equal(t, "@testuser great work!", event.Message.ArgumentText)
	assert.Equal(t, "Test User", event.Message.Sender.DisplayName)
	assert.Equal(t, "spaces/test-space", event.Space.Name)
}

func TestKudosStructure(t *testing.T) {
	// Test Kudos struct
	kudos := &Kudos{
		Command:     KudosCommand,
		UserID:      "123456789",
		Username:    "testuser",
		Description: "great work on the project",
	}
	
	assert.Equal(t, KudosCommand, kudos.Command)
	assert.Equal(t, "/kudos", string(kudos.Command))
	assert.Equal(t, "123456789", kudos.UserID)
	assert.Equal(t, "testuser", kudos.Username)
	assert.Equal(t, "great work on the project", kudos.Description)
}

func TestChatMessageResponse(t *testing.T) {
	// Test that we can create chat.Message responses properly
	message := &chat.Message{
		Text: "🎉 Kudos to <users/123456789> for great work!\n\nThey now have **5** total kudos.",
	}
	
	assert.NotNil(t, message)
	assert.Contains(t, message.Text, "🎉 Kudos")
	assert.Contains(t, message.Text, "<users/123456789>")
	assert.Contains(t, message.Text, "**5**")
}

func TestCommandConstants(t *testing.T) {
	// Test that constants are defined correctly
	assert.Equal(t, "/kudos", string(KudosCommand))
	
	// Test Commands type
	var cmd Commands = KudosCommand
	assert.Equal(t, "/kudos", string(cmd))
}

func TestErrorTypes(t *testing.T) {
	// Test that error types are defined
	assert.NotNil(t, invalidCommandError)
	assert.Equal(t, "Invalid command format", invalidCommandError.Error())
}