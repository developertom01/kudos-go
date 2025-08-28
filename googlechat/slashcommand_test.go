package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, result)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommandText(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.UserID)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isKudosCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}