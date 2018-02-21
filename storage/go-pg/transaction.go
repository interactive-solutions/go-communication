package gopg

import (
	"github.com/go-pg/pg"
	"github.com/interactive-solutions/go-communication"
)

func NewTransactionRepository(db *pg.DB) communication.TransactionRepository {
	return &transactionRepo{
		db: db,
	}
}

type transactionRepo struct {
	db *pg.DB
}

func (repo *transactionRepo) GetPending() ([]communication.Job, error) {
	var jobs []communication.Job

	if err := repo.db.Model(&jobs).Where("send_at is null").Select(); err != nil {
		if err == pg.ErrNoRows {
			return jobs, nil
		}

		return jobs, err
	}

	return jobs, nil
}

func (repo *transactionRepo) Create(job *communication.Job) error {
	return repo.db.Insert(job)
}

func (repo *transactionRepo) Update(job *communication.Job) error {
	return repo.db.Update(job)
}
