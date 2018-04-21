package gopg

import (
	"github.com/go-pg/pg"
	"github.com/interactive-solutions/go-communication"
)

func NewTransactionRepository(db *pg.DB) communication.JobRepository {
	return &jobRepository{
		db: db,
	}
}

type jobWrapper struct {
	TableName struct{} `sql:"communication_jobs, alias:cj" json:"-"`

	*communication.Job
}

type jobRepository struct {
	db *pg.DB
}

func (repo *jobRepository) Create(job *communication.Job) error {
	return repo.db.Insert(&jobWrapper{Job: job})
}

func (repo *jobRepository) Update(job *communication.Job) error {
	return repo.db.Update(&jobWrapper{Job: job})
}

func (repo *jobRepository) GetPending() ([]communication.Job, error) {
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

func (repo *jobRepository) Matching(criteria communication.JobCriteria) ([]communication.Job, int, error) {
	var jobs []communication.Job
	var wrappedJobs []jobWrapper

	builder := repo.db.Model(&wrappedJobs).
		Offset(criteria.Offset).
		Limit(criteria.Limit)

	if criteria.Type != "" {
		builder.Where("type = ?", criteria.Type)
	}

	if criteria.TemplateId != "" {
		builder.Where("template_id like ?", criteria.TemplateId+"%")
	}

	if criteria.Locale != "" {
		builder.Where("LOWER(locale) = LOWER(?)", criteria.Locale)
	}

	if criteria.Target != "" {
		builder.Where("LOWER(target) = LOWER(?)", criteria.Target+"%")
	}

	if criteria.ExternalId != "" {
		builder.Where("external_id = ?", criteria.ExternalId)
	}

	if !criteria.SentAfter.IsZero() {
		builder.Where("sent_at >= ?", criteria.SentAfter)
	}

	if !criteria.SentBefore.IsZero() {
		builder.Where("sent_at <= ?", criteria.SentBefore)
	}

	for col, dir := range criteria.Sorting {
		builder.Order("%s %s", col, dir)
	}

	count, err := builder.SelectAndCount()
	if err != nil && err != pg.ErrNoRows {
		return jobs, 0, err
	}

	for _, job := range wrappedJobs {
		jobs = append(jobs, *job.Job)
	}

	return jobs, count, nil
}
