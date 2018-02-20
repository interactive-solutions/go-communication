package gopg

import (
	"github.com/interactive-solutions/go-communication"
	"gopkg.in/pg.v5"
)

func NewTemplateRepository(db *pg.DB) communication.TemplateRepository {
	return &templateRepository{
		db: db,
	}
}

type templateRepository struct {
	db *pg.DB
}

func (repo *templateRepository) Get(id, locale string) (communication.Template, error) {
	template := communication.Template{}

	if err := repo.db.Model(&template).Where("template_id = ? AND locale = ?", id, locale).Select(); err != nil {
		if err == pg.ErrNoRows {
			return template, communication.TemplateNotFoundErr
		}

		return template, err
	}

	return template, nil
}

func (repo *templateRepository) Create(template *communication.Template) error {
	return repo.db.Insert(template)
}

func (repo *templateRepository) Update(template *communication.Template) error {
	return repo.db.Update(template)
}

func (repo *templateRepository) Delete(template *communication.Template) error {
	return repo.db.Delete(template)
}
