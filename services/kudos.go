package services

import (
	"time"

	"github.com/developertom01/go-kudos/data"
)

type Platform string

const (
	SlackPlatform Platform = "slack"
)

type (
	KudosService struct {
	}

	KudosResponse struct {
		Total       int64     `json:"total"`
		Description string    `json:"description"`
		Username    string    `json:"username"`
		From        string    `json:"from"`
		CreatedAt   time.Time `json:"updated_at,omitempty"`
		Platform    Platform  `json:"platform"`
	}

	KudosPayload struct {
		OrganizationId string `json:"organization_id"`
		ToUsername       string `json:"to_user_name"`
		Description    string `json:"description"`
		InstallationId string `json:"installation_id"`
		FromUsername    string `json:"from_user_name"`
	}
)

func NewKudosService() *KudosService {
	return &KudosService{}
}

func (kudosService *KudosService) HandleKudos(payload KudosPayload, database *data.Database) (*KudosResponse, error) {
	kudus, err := database.CreateKudos(
		payload.FromUsername,
		payload.ToUsername,
		payload.Description,
		payload.InstallationId,
	)
	if err != nil {
		return nil, err
	}

	kudusCount, err := database.GetKudusCountForUser(
		payload.InstallationId,
		payload.ToUsername,
	)

	if err != nil {
		return nil, err
	}

	kudosResponse := &KudosResponse{
		Total:       kudusCount,
		Description: kudus.Description,
		Username:    kudus.ToUser.Username,
		From:        kudus.FromUser.Username,
		CreatedAt:   kudus.CreatedAt,
		Platform:    Platform(kudus.Installation.Platform),
	}

	return kudosResponse, nil
}
