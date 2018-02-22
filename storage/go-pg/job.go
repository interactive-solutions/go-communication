package gopg

import (
	"github.com/go-pg/pg"
	"github.com/interactive-solutions/go-communication"
)

func NewTransactionRepository(db *pg.DB) communication.JobRepository {
	return &transactionRepo{
		db: db,
	}
}

type jobWrapper struct {
	TableName struct{} `sql:"communication_jobs, alias:cj" json:"-"`

	*communication.Job
}

type transactionRepo struct {
	db *pg.DB
}

func (repo *transactionRepo) GetPending() ([]communication.Job, error) {
	var jobs []communication.Job
	var wrappedJobs []jobWrapper

	if err := repo.db.Model(&wrappedJobs).Where("sent_at is null").Select(); err != nil {
		if err == pg.ErrNoRows {
			return jobs, nil
		}

		return jobs, err
	}

	for _, j := range wrappedJobs {
		jobs = append(jobs, *j.Job)
	}

	return jobs, nil
}

func (repo *transactionRepo) Create(job *communication.Job) error {
	return repo.db.Insert(&jobWrapper{Job: job})
}

func (repo *transactionRepo) Update(job *communication.Job) error {
	return repo.db.Update(&jobWrapper{Job: job})
}
