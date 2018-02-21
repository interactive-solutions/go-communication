package gopg

import (
	"github.com/go-pg/pg"
	"github.com/interactive-solutions/go-communication"
)

func NewTemplateRepository(db *pg.DB) communication.TemplateRepository {
	return &templateRepository{
		db: db,
	}
}

type templateRepository struct {
	db *pg.DB
}

type templateWrapper struct {
	TableName struct{} `sql:"communication_templates,alias:ct" json:"-"`

	*communication.Template
}

func (repo *templateRepository) Get(id, locale string) (communication.Template, error) {

	wrapped := &templateWrapper{
		Template: &communication.Template{},
	}

	if err := repo.db.Model(wrapped).Where("template_id = ? AND locale = ?", id, locale).Select(); err != nil {
		if err == pg.ErrNoRows {
			return *wrapped.Template, communication.TemplateNotFoundErr
		}

		return *wrapped.Template, err
	}

	return *wrapped.Template, nil
}

func (repo *templateRepository) Create(template *communication.Template) error {
	return repo.db.Insert(&templateWrapper{Template: template})
}

func (repo *templateRepository) Update(template *communication.Template) error {
	return repo.db.Update(&templateWrapper{Template: template})
}

func (repo *templateRepository) Delete(template *communication.Template) error {
	return repo.db.Delete(&templateWrapper{Template: template})
}
