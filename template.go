package communication

import "time"

type Template struct {
	TemplateId string `sql:",pk" json:"id"`
	Locale     string `sql:",pk" json:"locale"`

	Enabled     bool   `sql:",notnull" json:"enabled"`
	Description string `sql:",notnull" json:"description"`

	Parameters       map[string]interface{} `json:"parameters"`
	UpdateParameters bool                   `sql:",notnull" json:"updateParameters"`

	Subject  string `json:"subject"`
	TextBody string `json:"textBody"`
	HtmlBody string `json:"htmlBody"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TemplateService interface {
	Render(id, locale string, parameters map[string]interface{}) error
}
