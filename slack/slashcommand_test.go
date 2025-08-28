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
			name:  "Valid Slack @mention format",
			input: "/kudos <@U1234567890> great work on the project",
			expected: &Kudos{
				Command:     KudosCommand,
				UserID:      "U1234567890",
				Description: "great work on the project",
			},
			shouldError: false,
		},
		{
			name:  "Valid legacy @username format",
			input: "/kudos @john awesome debugging skills",
			expected: &Kudos{
				Command:     KudosCommand,
				Username:    "john",
				Description: "awesome debugging skills",
			},
			shouldError: false,
		},
		{
			name:  "Multi-word description",
			input: "/kudos @jane thank you for helping with the complex database optimization task",
			expected: &Kudos{
				Command:     KudosCommand,
				Username:    "jane",
				Description: "thank you for helping with the complex database optimization task",
			},
			shouldError: false,
		},
		{
			name:        "Missing description",
			input:       "/kudos @john",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "Missing username",
			input:       "/kudos great work",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "Invalid command",
			input:       "/other @john great work",
			expected:    nil,
			shouldError: true,
		},
		{
			name:        "No @ prefix",
			input:       "/kudos john great work",
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

func TestParseCommandTextSlackMentionFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected UserID
	}{
		{
			name:     "Standard Slack user ID",
			input:    "/kudos <@U1234567890> great work",
			expected: "U1234567890",
		},
		{
			name:     "Slack user ID with different format",
			input:    "/kudos <@W9876543210> excellent debugging",
			expected: "W9876543210",
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
			input:       "/kudos @alice Thank you for the excellent code review! Your feedback was spot-on.",
			expectedDesc: "Thank you for the excellent code review! Your feedback was spot-on.",
			shouldError: false,
		},
		{
			name:        "Description with emojis",
			input:       "/kudos <@U1234567890> Amazing work on the deployment! 🚀🎉",
			expectedDesc: "Amazing work on the deployment! 🚀🎉",
			shouldError: false,
		},
		{
			name:        "Multiple spaces in description normalized",
			input:       "/kudos @bob    Excellent    debugging    skills",
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