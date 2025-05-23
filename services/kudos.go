package services

type Platform string

const (
	SlackPlatform Platform = "slack"
)

type (
	KudosService struct {
	}

	KudosResponse struct {
		Total          int      `json:"total"`
		Description    string   `json:"description"`
		Username       string   `json:"username"`
		From           string   `json:"from"`
		UpdatedAt      string   `json:"updated_at"`
		OrganizationId string   `json:"organization_id"`
		Platform       Platform `json:"platform"`
	}

	KudosPayload struct {
		OrganizationId string `json:"organization_id"`
		Username       string `json:"username"`
		Description    string `json:"description"`
		Platform       Platform `json:"platform"`
	}
)

func NewKudosService() *KudosService {
	return &KudosService{}
}

func (kudosService *KudosService) HandleKudos(payload KudosPayload) (*KudosResponse, error) {
	return nil, nil
}
