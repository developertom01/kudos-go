package main

import (
	"github.com/developertom01/go-kudos/services"
	"github.com/developertom01/go-kudos/slack/config"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

func main() {
	services := services.NewKudosService()
	slackApi := slack.New(config.SLACK_API_TOKEN)

	r := gin.Default()
	r.GET(config.KUDOS_SLASH_COMMAND, func(c *gin.Context) {
		slashCommand, err := slack.SlashCommandParse(c.Request)

		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request",
			})
			return
		}

		err = handleSlashCommand(slashCommand, services, slackApi)

		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request",
			})
			return
		}
		c.JSON(200, nil)
	})

	r.Run(config.PORT)

}
