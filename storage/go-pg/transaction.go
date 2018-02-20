package gopg

import (
	"github.com/interactive-solutions/go-communication"
	"gopkg.in/pg.v5"
)

type transaction struct {
	db *pg.DB
}

func NewTransactionRepository(db *pg.DB) communication.TransactionRepository {
	return &transaction{
		db: db,
	}
}


