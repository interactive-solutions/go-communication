package gopg

import (
	"github.com/go-pg/pg/types"
	"time"

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
	template.UpdatedAt = time.Now()

	return repo.db.Update(&templateWrapper{Template: template})
}

func (repo *templateRepository) Delete(template *communication.Template) error {
	return repo.db.Delete(&templateWrapper{Template: template})
}

func (repo *templateRepository) Matching(criteria communication.TemplateCriteria) ([]communication.Template, int, error) {
	var wrapped []templateWrapper
	templates := make([]communication.Template, 0)

	builder := repo.db.Model(&wrapped).
		Offset(criteria.Offset).
		Limit(criteria.Limit)

	if criteria.TemplateId != "" {
		builder.Where("template_id like ?", criteria.TemplateId+"%")
	}

	if criteria.Locale != "" {
		builder.Where("LOWER(locale) = LOWER(?)", criteria.Locale)
	}

	if criteria.Subject != "" {
		builder.Where("LOWER(subject) = LOWER(?)", criteria.Subject+"%")
	}

	if !criteria.UpdatedAfter.IsZero() {
		builder.Where("updated_at >= ?", criteria.UpdatedAfter)
	}

	if !criteria.UpdatedBefore.IsZero() {
		builder.Where("updated_at <= ?", criteria.UpdatedBefore)
	}

	for col, dir := range criteria.Sorting {
		builder.OrderExpr("%s %s", types.F(col), types.Q(dir))
	}

	count, err := builder.SelectAndCount()
	if err != nil && err != pg.ErrNoRows {
		return templates, 0, err
	}

	for _, t := range wrapped {
		templates = append(templates, *t.Template)
	}

	return templates, count, nil
}
