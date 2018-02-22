package communication

import "github.com/pkg/errors"

var (
	TemplateNotFoundErr = errors.New("The template was not found")
	JobNotFoundErr      = errors.New("The transaction was not found")
)

type TemplateRepository interface {
	Get(id, locale string) (Template, error)
	GetAll() ([]Template, error)

	Create(template *Template) error
	Update(template *Template) error
	Delete(template *Template) error
}

type JobRepository interface {
	GetPending() ([]Job, error)

	Create(*Job) error
	Update(*Job) error
}
