package communication

type Template struct {
	TemplateId string `pg:",pk" json:"id"`
	Locale     string `pg:",pk" json:"locale"`

	Enabled     bool
	Description string `json:"description"`

	Parameters       map[string]interface{} `json:"parameters"`
	UpdateParameters bool                   `json:"updateParameters"`

	Subject  string `json:"subject"`
	TextBody string `json:"textBody"`
	HtmlBody string `json:"htmlBody"`
}

type TemplateService interface {
	Render(id, locale string, parameters map[string]interface{}) error
}
