package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/developertom01/go-kudos/data"
	"github.com/developertom01/go-kudos/services"
	"github.com/slack-go/slack"
)

type Commands string

const (
	KudosCommand Commands = "/kudos"
)

var (
	invalidCommandError = errors.New("Invalid command format")
)

type Kudos struct {
	Command     Commands
	UserID      string  // Slack user ID from @mention
	Username    string  // Resolved username
	Description string  // Full description text
}

// eg. /kudos <@U1234567890> kudos for great work
// or /kudos @username kudos for great work
func parseCommandText(text string) (*Kudos, error) {
	// Split the text into parts
	parts := strings.Fields(text)

	if len(parts) < 3 {
		return nil, errors.New("command format: /kudos @user description")
	}

	if parts[0] != string(KudosCommand) {
		return nil, invalidCommandError
	}

	userPart := parts[1]
	description := strings.Join(parts[2:], " ")

	kudos := &Kudos{
		Command:     KudosCommand,
		Description: description,
	}

	// Check if it's a Slack user mention format <@U1234567890>
	slackMentionRegex := regexp.MustCompile(`^<@([A-Z0-9]+)>$`)
	if matches := slackMentionRegex.FindStringSubmatch(userPart); len(matches) > 1 {
		kudos.UserID = matches[1]
		// Username will be resolved via Slack API
	} else if strings.HasPrefix(userPart, "@") {
		// Legacy @username format
		kudos.Username = strings.TrimPrefix(userPart, "@")
	} else {
		return nil, errors.New("user must be mentioned with @ or Slack @mention format")
	}

	return kudos, nil
}

func handleSlashCommand(slashCommand slack.SlashCommand, service *services.KudosService, slackApi *slack.Client, database *data.Database) error {
	// Get installation for this team to use the correct token
	installation, err := database.GetInstallationByTeamID(slashCommand.TeamID)
	if err != nil {
		return errors.New("App not installed for this workspace")
	}
	
	// Create client with the installation's bot token
	installedSlackApi := slack.New(installation.BotUserOAuthToken)
	
	kudos, err := parseCommandText(slashCommand.Text)
	if err != nil {
		return err
	}

	// Resolve Slack user ID to username if needed
	if kudos.UserID != "" {
		user, err := installedSlackApi.GetUserInfo(kudos.UserID)
		if err != nil {
			return fmt.Errorf("failed to resolve user: %v", err)
		}
		kudos.Username = user.Name
	}

	var orgId = slashCommand.EnterpriseID
	if orgId == "" {
		orgId = slashCommand.TeamID
	}

	kudosPayload := services.KudosPayload{
		OrganizationId: orgId,
		ToUsername:     kudos.Username,
		Description:    kudos.Description,
		InstallationId: slashCommand.APIAppID,
		FromUsername:   slashCommand.UserName,
	}

	kudosResponse, err := service.HandleKudos(kudosPayload, database)
	if err != nil {
		return err
	}

	// Create the @mention format for the response
	var userMention string
	if kudos.UserID != "" {
		userMention = fmt.Sprintf("<@%s>", kudos.UserID)
	} else {
		userMention = fmt.Sprintf("@%s", kudos.Username)
	}

	// Send the response back to Slack using the installation-specific client
	_, _, err = installedSlackApi.PostMessage(slashCommand.ChannelID, 
		slack.MsgOptionText(
			fmt.Sprintf("Kudos to %s for %s! 🎉\nThey now have %d total kudos.", 
				userMention, kudos.Description, kudosResponse.Total), 
			false),
		slack.MsgOptionAsUser(false),
		slack.MsgOptionIconEmoji(":tada:"),
	)

	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}

	return nil
}
