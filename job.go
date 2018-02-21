package communication

import (
	"time"

	"github.com/satori/go.uuid"
)

type JobType uint

const (
	JobSms JobType = iota
	JobEmail
)

type Job struct {
	Uuid uuid.UUID `pg:",pk" json:"uuid"`
	Type JobType   `json:"type"`

	TemplateId string `json:"templateId"`
	Locale     string `json:"locale"`
	Target     string `json:"target"`

	Params map[string]interface{} `json:"params"`

	SentAt    *time.Time `json:"sentAt"`
	CreatedAt time.Time  `json:"createdAt"`
}
