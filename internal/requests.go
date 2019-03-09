package internal

type TestTemplateRequest struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Target string `json:"target"`
}

type UpdateTemplateRequest struct {
	UpdateParameters bool   `json:"updateParameters"`
	Enabled          bool   `json:"enabled"`
	Description      string `json:"description"`

	Subject  string `json:"subject"`
	HtmlBody string `json:"htmlBody"`
	TextBody string `json:"textBody"`
}

type ResubscribeRequest struct {
	Email     string   `json:"email"`
	Templates []string `json:"templates"`
}
