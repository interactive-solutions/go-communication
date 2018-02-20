package gopg

import (
	"github.com/interactive-solutions/go-communication"
	"gopkg.in/pg.v5"
)

type templateRepository struct {
	db *pg.DB
}

func NewTemplateRepository(db *pg.DB) communication.TemplateRepository {
	return &templateRepository{
		db: db,
	}
}
