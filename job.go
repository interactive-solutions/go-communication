package communication

import (
	"time"

	"github.com/satori/go.uuid"
)

type JobType string

const (
	JobSms   JobType = "sms"
	JobEmail JobType = "email"
)

type Job struct {
	Uuid       uuid.UUID `sql:",pk" json:"uuid"`
	ExternalId string    `sql:",notnull" json:"externalId"`
	Type       JobType   `json:"type"`

	TemplateId string `json:"templateId"`
	Locale     string `json:"locale"`
	Target     string `json:"target"`

	Params map[string]interface{} `json:"params"`

	SentAt    *time.Time `json:"sentAt"`
	CreatedAt time.Time  `json:"createdAt"`
}
