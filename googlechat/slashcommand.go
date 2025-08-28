package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"
)

type Commands string

const (
	KudosCommand Commands = "/kudos"
)

var (
	invalidCommandError = errors.New("Invalid command format")
)

// GoogleChatEvent represents a Google Chat event
type GoogleChatEvent struct {
	Type    string `json:"type"`
	EventTime string `json:"eventTime"`
	Message struct {
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
	} `json:"message"`
	Space struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"space"`
	User struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
	} `json:"user"`
}

type Kudos struct {
	Command     Commands
	UserID      string  // Google Chat user ID from @mention
	Username    string  // Resolved username
	Description string  // Full description text
}

// parseGoogleChatMention parses Google Chat @mention format
// Google Chat uses format like <users/USER_ID> or @username
func parseCommandText(text string) (*Kudos, error) {
	// Remove leading slash if present and split into parts
	text = strings.TrimPrefix(text, "/kudos")
	text = strings.TrimSpace(text)
	
	parts := strings.Fields(text)

	if len(parts) < 2 {
		return nil, errors.New("command format: /kudos @user description")
	}

	userPart := parts[0]
	description := strings.Join(parts[1:], " ")

	kudos := &Kudos{
		Command:     KudosCommand,
		Description: description,
	}

	// Check if it's a Google Chat user mention format <users/USER_ID>
	googleMentionRegex := regexp.MustCompile(`^<users/([^>]+)>$`)
	if matches := googleMentionRegex.FindStringSubmatch(userPart); len(matches) > 1 {
		kudos.UserID = matches[1]
		// Username will be resolved via Google Chat API
	} else if strings.HasPrefix(userPart, "@") {
		// Legacy @username format
		kudos.Username = strings.TrimPrefix(userPart, "@")
	} else {
		return nil, errors.New("user must be mentioned with @ or Google Chat @mention format")
	}

	return kudos, nil
}

func handleGoogleChatCommand(event GoogleChatEvent, service *services.KudosService, database *data.Database) (*chat.Message, error) {
	// Extract team/space ID from the space name
	spaceID := event.Space.Name
	
	if spaceID == "" {
		log.Printf("Missing space ID in event")
		return nil, errors.New("Invalid space information")
	}
	
	// Get installation for this space to use the correct token
	installation, err := database.GetInstallationByTeamID(spaceID)
	if err != nil {
		log.Printf("Installation not found for space %s: %v", spaceID, err)
		return nil, errors.New("App not installed for this Google Chat space. Please visit /auth/googlechat to install.")
	}
	
	// Parse the command text
	kudos, err := parseCommandText(event.Message.ArgumentText)
	if err != nil {
		log.Printf("Command parsing error: %v", err)
		return nil, fmt.Errorf("❌ %s\n\nUsage: `/kudos @user description` or `/kudos <users/USER_ID> description`", err.Error())
	}

	// Resolve Google Chat user ID to username if needed
	if kudos.UserID != "" {
		log.Printf("Resolving user ID: %s", kudos.UserID)
		
		// Create OAuth2 token from stored tokens
		token := &oauth2.Token{
			AccessToken:  installation.AccessToken,
			RefreshToken: installation.BotUserOAuthToken, // We stored refresh token here
		}
		
		// Create Chat service with the installation's token
		ctx := context.Background()
		oauthConfig := &oauth2.Config{}
		client := oauthConfig.Client(ctx, token)
		chatService, err := chat.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Printf("Failed to create chat service: %v", err)
			// Fall back to using user ID as username
			kudos.Username = kudos.UserID
		} else {
			// Try to get user info - in Google Chat, this might not be directly available
			// For now, we'll use the user ID as the username
			kudos.Username = kudos.UserID
			// Suppress unused variable warning
			_ = chatService
		}
	}

	if kudos.Username == "" {
		return nil, errors.New("❌ Unable to resolve user information")
	}

	// Extract organization ID from space
	var orgId = spaceID
	
	// Extract sender info
	senderName := event.Message.Sender.DisplayName
	if senderName == "" {
		senderName = event.Message.Sender.Name
	}
	
	if senderName == "" {
		log.Printf("Missing sender information in event")
		return nil, errors.New("❌ Unable to identify sender")
	}

	log.Printf("Processing kudos: from=%s, to=%s, description=%s", senderName, kudos.Username, kudos.Description)

	kudosPayload := services.KudosPayload{
		OrganizationId: orgId,
		ToUsername:     kudos.Username,
		Description:    kudos.Description,
		InstallationId: spaceID,
		FromUsername:   senderName,
	}

	kudosResponse, err := service.HandleKudos(kudosPayload, database)
	if err != nil {
		log.Printf("Kudos service error: %v", err)
		return nil, fmt.Errorf("❌ Failed to process kudos: %s", err.Error())
	}

	// Create the @mention format for the response
	var userMention string
	if kudos.UserID != "" {
		userMention = fmt.Sprintf("<users/%s>", kudos.UserID)
	} else {
		userMention = fmt.Sprintf("@%s", kudos.Username)
	}

	// Create response message for Google Chat
	responseText := fmt.Sprintf("🎉 Kudos to %s for %s!\n\nThey now have **%d** total kudos.", 
		userMention, kudos.Description, kudosResponse.Total)

	log.Printf("Kudos processed successfully: total=%d", kudosResponse.Total)

	return &chat.Message{
		Text: responseText,
	}, nil
}