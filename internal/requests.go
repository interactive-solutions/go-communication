package internal

type UpdateTemplateRequest struct {
	UpdateParameters bool   `json:"updateParameters"`
	Enabled          bool   `json:"enabled"`
	Description      string `json:"description"`

	Subject  string `json:"subject"`
	HtmlBody string `json:"htmlBody"`
	TextBody string `json:"textBody"`
}
