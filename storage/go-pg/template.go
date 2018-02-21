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
	*communication.Template

	TableName struct{} `sql:"communication_templates,alias:ct" json:"-"`
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
	return repo.db.Insert(&templateWrapper{Template: template})
}

func (repo *templateRepository) Update(template *communication.Template) error {
	return repo.db.Update(&templateWrapper{Template: template})
}

func (repo *templateRepository) Delete(template *communication.Template) error {
	return repo.db.Delete(&templateWrapper{Template: template})
}
