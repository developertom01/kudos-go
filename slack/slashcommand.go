package main

import (
	"errors"
	"fmt"
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
	Username    string
	Description *string
}

// eg. /kudos @user kudos for great work
func parseCommandText(text string) (*Kudos, error) {
	// Split the text into parts
	parts := strings.Split(text, " ")

	if len(parts) < 2 {
		return nil, invalidCommandError
	}

	if parts[0] == string(KudosCommand) {
		return &Kudos{
			Command:     KudosCommand,
			Username:    parts[1],
			Description: &parts[2],
		}, nil
	}

	return nil, invalidCommandError
}

func handleSlashCommand(slashCommand slack.SlashCommand, service *services.KudosService, slackApi *slack.Client, database *data.Database) error {
	kudos, err := parseCommandText(slashCommand.Text)

	if err != nil {
		return err
	}

	kudos.Username = strings.TrimPrefix(kudos.Username, "@")

	var orgId = slashCommand.EnterpriseID
	if orgId == "" {
		orgId = slashCommand.TeamID
	}

	kudosPayload := services.KudosPayload{
		OrganizationId: orgId,
		ToUsername:       kudos.Username,
		Description:    *kudos.Description,
		InstallationId: slashCommand.APIAppID,
		FromUsername:     slashCommand.UserName,
	}

	kudosResponse, err := service.HandleKudos(kudosPayload, database)

	// Send the response back to Slack as thread

	text, atch, err := slackApi.PostMessage(slashCommand.ChannelID, slack.MsgOptionText(
		fmt.Sprintf("Kudos to %s for %s", kudos.Username, *kudos.Description), false),
		slack.MsgOptionText(fmt.Sprintf("You now have %d kudos", kudosResponse.Total), false),
		slack.MsgOptionTS(slashCommand.TriggerID),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionIconEmoji(":tada:"),
		slack.MsgOptionReplaceOriginal(slashCommand.ResponseURL),
		slack.MsgOptionMeMessage(),
	)

	if err != nil {

		return err
	}

	fmt.Println("kudosResponse", text, atch)

	return err
}
