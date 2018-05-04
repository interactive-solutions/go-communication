package communication

import (
	"encoding/base64"
	"html/template"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestApplication(t *testing.T) {
	suite.Run(t, new(applicationTestSuite))
}

type applicationTestSuite struct {
	suite.Suite
}

func (suite *applicationTestSuite) TestRenderWithCustomFunc() {
	funcMap := template.FuncMap{
		"base64": func(test interface{}) string {
			var str string
			switch test.(type) {
			case int, int8, int16, int32, int64:
				str = strconv.FormatInt(int64(test.(int)), 10)

			case uint, uint8, uint16, uint32, uint64:
				str = strconv.FormatUint(uint64(test.(uint64)), 10)

			case string:
				str = test.(string)
			}

			return base64.URLEncoding.EncodeToString([]byte(str))
		},
	}

	app, err := NewApplication(
		SetTemplateFuncMap(funcMap),
		SetJobRepo(&jobRepository{}),
		SetTemplateRepo(&templateRepository{}),
	)

	if !assert.NoError(suite.T(), err, "Failed to create the new application") {
		return
	}

	tpl := Template{
		Subject:  "hello world",
		TextBody: "text body",
		HtmlBody: "html body https://interactivesolutions.se?ref={{ .id | base64 }}",
		Parameters: map[string]interface{}{
			"id": 100,
		},
	}

	subject, text, html, err := app.(*application).Render(tpl)
	if !assert.NoError(suite.T(), err, "Failed to render the template") {
		return
	}

	assert.Equal(suite.T(), "hello world", subject)
	assert.Equal(suite.T(), "text body", text)
	assert.Equal(suite.T(), "html body https://interactivesolutions.se?ref=MTAw", html)
}

type templateRepository struct {
	GetTemplate    Template
	MatchTemplates []Template
}

func (repo *templateRepository) Get(id, locale string) (Template, error) {
	return repo.GetTemplate, nil
}

func (repo *templateRepository) Matching(criteria TemplateCriteria) ([]Template, int, error) {
	return repo.MatchTemplates, len(repo.MatchTemplates), nil
}

func (repo *templateRepository) Create(template *Template) error {
	return nil
}

func (repo *templateRepository) Update(template *Template) error {
	return nil
}

func (repo *templateRepository) Delete(template *Template) error {
	return nil
}

type jobRepository struct {
	PendingJobs  []Job
	MatchingJobs []Job
}

func (repo *jobRepository) GetPending() ([]Job, error) {
	return repo.PendingJobs, nil
}

func (repo *jobRepository) Matching(criteria JobCriteria) ([]Job, int, error) {
	return repo.MatchingJobs, len(repo.MatchingJobs), nil
}

func (repo *jobRepository) Create(*Job) error {
	return nil
}

func (repo *jobRepository) Update(*Job) error {
	return nil
}
