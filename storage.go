package communication

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	TemplateNotFoundErr = errors.New("The template was not found")
	JobNotFoundErr      = errors.New("The transaction was not found")
)

var templateSortingMap = map[string]string{
	"enabled":    "enabled",
	"updatedAt":  "updated_at",
	"createdAt":  "created_at",
	"templateId": "template_id",
}

var jobSortingMap = map[string]string{
	"sentAt":    "sent_at",
	"createdAt": "created_at",
}

type TemplateCriteria struct {
	Offset int
	Limit  int

	Locale     string
	TemplateId string
	Subject    string

	UpdatedAfter  time.Time
	UpdatedBefore time.Time

	Sorting map[string]string
}

func PopulateTemplateCriteria(r *http.Request) TemplateCriteria {
	criteria := TemplateCriteria{
		Offset:  0,
		Limit:   10,
		Sorting: map[string]string{},
	}

	criteria.Locale = r.FormValue("locale")
	criteria.Subject = r.FormValue("subject")
	criteria.TemplateId = r.FormValue("templateId")

	if after, err := time.Parse(time.RFC3339, r.FormValue("updatedAfter")); err == nil {
		criteria.UpdatedAfter = after
	}

	if before, err := time.Parse(time.RFC3339, r.FormValue("updatedBefore")); err == nil {
		criteria.UpdatedBefore = before
	}

	if limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64); err == nil {
		criteria.Limit = int(limit)
	}

	if offset, err := strconv.ParseInt(r.FormValue("offset"), 10, 64); err == nil {
		criteria.Offset = int(offset)
	}

	if sorting := r.FormValue("sorting"); sorting != "" {
		sorts := strings.Split(sorting, ",")

		for _, sort := range sorts {
			split := strings.Split(sort, ":")
			// Remove invalid splits
			if len(split) != 2 || (split[1] != "asc" && split[1] != "desc") {
				continue
			}

			// Only allow sorting on specific fields
			if column, ok := templateSortingMap[split[0]]; ok {
				criteria.Sorting[column] = split[1]
			}
		}
	} else {
		criteria.Sorting["template_id"] = "desc"
		criteria.Sorting["locale"] = "asc"
	}

	return criteria
}

type TemplateRepository interface {
	Get(id, locale string) (Template, error)
	Matching(criteria TemplateCriteria) ([]Template, int, error)

	Create(template *Template) error
	Update(template *Template) error
	Delete(template *Template) error
}

type JobCriteria struct {
	Limit  int
	Offset int

	Type       string
	TemplateId string
	Locale     string
	Target     string
	ExternalId string

	SentAfter  time.Time
	SentBefore time.Time

	Sorting map[string]string
}

func PopulateJobCriteria(r *http.Request) JobCriteria {
	criteria := JobCriteria{
		Offset:  0,
		Limit:   10,
		Sorting: map[string]string{},
	}

	criteria.Type = r.FormValue("type")
	criteria.Locale = r.FormValue("locale")
	criteria.TemplateId = r.FormValue("templateId")

	if after, err := time.Parse(time.RFC3339, r.FormValue("sentAfter")); err == nil {
		criteria.SentAfter = after
	}

	if before, err := time.Parse(time.RFC3339, r.FormValue("sentBefore")); err == nil {
		criteria.SentBefore = before
	}

	if limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64); err == nil {
		criteria.Limit = int(limit)
	}

	if offset, err := strconv.ParseInt(r.FormValue("offset"), 10, 64); err == nil {
		criteria.Offset = int(offset)
	}

	if sorting := r.FormValue("sorting"); sorting != "" {
		sorts := strings.Split(sorting, ",")

		for _, sort := range sorts {
			split := strings.Split(sort, ":")
			// Remove invalid splits
			if len(split) != 2 || (split[1] != "asc" && split[1] != "desc") {
				continue
			}

			// Only allow sorting on specific fields
			if column, ok := jobSortingMap[split[0]]; ok {
				criteria.Sorting[column] = split[1]
			}
		}
	}

	return criteria
}

type JobRepository interface {
	GetPending() ([]Job, error)
	Matching(criteria JobCriteria) ([]Job, int, error)

	Create(*Job) error
	Update(*Job) error
}
