package communication

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/pkg/errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const UserAgent = "InteractiveSolutions/GoCommunication-1.0"

type Application interface {
	HttpHandler() *HttpHandler
	SendEmail(id, locale, email, externalId string, params map[string]interface{}) error
	SendSms(id, locale, number, externalId string, params map[string]interface{}) error
	Shutdown(ctx context.Context)
}

type AppOption func(a *application)
type RenderFunc func(body string, params map[string]interface{}) (string, error)

func SetFallbackLocale(locale string) AppOption {
	return func(a *application) {
		a.fallbackLocale = locale
	}
}

func SetDefaultSmsTransport(transport Transport) AppOption {
	return func(a *application) {
		a.defaultSmsTransport = transport
	}
}

func SetDefaultEmailTransport(transport Transport) AppOption {
	return func(a *application) {
		a.defaultEmailTransport = transport
	}
}

func SetTemplateRepo(repo TemplateRepository) AppOption {
	return func(a *application) {
		a.templateRepo = repo
	}
}

func SetJobRepo(repo JobRepository) AppOption {
	return func(a *application) {
		a.jobRepo = repo
	}
}

func SetWorkerCount(count int) AppOption {
	return func(a *application) {
		a.workerCount = count
	}
}

func SetTemplateFuncMap(funcMap template.FuncMap) AppOption {
	return func(a *application) {
		a.templateFuncMap = funcMap
	}
}

func SetHtmlToTextConverter(f func (string) string) AppOption {
	return func (a *application) {
		a.htmlToTextConverter = f
	}
}

func SetLogger(logger logrus.FieldLogger) AppOption {
	return func (a *application) {
		a.logger = logger
	}
}

func SetStaticParams(params map[string]interface{}) AppOption {
	return func (a *application) {
		a.staticParams = params
	}
}

type application struct {
	logger logrus.FieldLogger

	workerCtx    context.Context
	workerCancel context.CancelFunc

	workerQueue chan *Job
	workerCount int

	templateRepo TemplateRepository
	jobRepo      JobRepository

	fallbackLocale        string
	defaultSmsTransport   Transport
	defaultEmailTransport Transport

	templateFuncMap template.FuncMap

	htmlToTextConverter func (string) string

	staticParams map[string]interface{}
}

func NewApplication(options ...AppOption) (Application, error) {
	app := &application{
		logger: logrus.New(),

		workerQueue: make(chan *Job, 1000),
		workerCount: 5,
	}

	for _, option := range options {
		option(app)
	}

	if err := app.ensureUsableConfiguration(); err != nil {
		return app, err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	app.workerCancel = cancel

	for i := 0; i <= app.workerCount; i++ {
		go app.worker(ctx)
	}

	jobs, err := app.jobRepo.GetPending()
	if err != nil {
		return app, err
	}

	for _, job := range jobs {
		cpy := job

		// Queue the copy of the join
		app.queue(&cpy)
	}

	return app, nil
}

func (a *application) HttpHandler() *HttpHandler {
	return &HttpHandler{
		app: a,
	}
}

func (a *application) SendEmail(id, locale, email, externalId string, params map[string]interface{}) error {
	if a.defaultEmailTransport == nil {
		return errors.New("No email transport configured")
	}

	job := &Job{
		Uuid:       uuid.New(),
		ExternalId: externalId,
		Type:       JobEmail,
		TemplateId: id,
		Locale:     locale,
		Target:     email,
		Params:     params,
		CreatedAt:  time.Now(),
	}

	if err := a.jobRepo.Create(job); err != nil {
		return err
	}

	a.queue(job)

	return nil
}

func (a *application) SendSms(id, locale, number, externalId string, params map[string]interface{}) error {
	if a.defaultSmsTransport == nil {
		return errors.New("No sms transport configured")
	}

	job := &Job{
		Uuid:       uuid.New(),
		ExternalId: externalId,
		Type:       JobSms,
		TemplateId: id,
		Locale:     locale,
		Target:     number,
		Params:     params,
		CreatedAt:  time.Now(),
	}

	if err := a.jobRepo.Create(job); err != nil {
		return err
	}

	a.queue(job)

	return nil
}

func (a *application) Shutdown(ctx context.Context) {
	<-ctx.Done()
	a.workerCancel()
}

func (a *application) ensureUsableConfiguration() error {
	if a.templateRepo == nil {
		return errors.New("Missing template repository")
	}

	if a.jobRepo == nil {
		return errors.New("Missing transaction repository")
	}

	return nil
}

func (a *application) queue(job *Job) {
	go func() {
		a.workerQueue <- job
	}()
}

func (a *application) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-a.workerQueue:
			if !ok {
				return
			}

			if err := a.process(job); err != nil {
				a.logger.
					WithField("job", job).
					WithError(err).
					Error("failed to process job")

				continue
			}

			now := time.Now()

			job.SentAt = &now

			if err := a.jobRepo.Update(job); err != nil {
				a.logger.
					WithField("job", job).
					WithError(err).
					Error("failed to update job in transaction repo")
			}
		}
	}
}

func (a *application) createMockTemplate(templateId, locale string) (Template, error) {
	tpl := Template{
		TemplateId:       templateId,
		Locale:           locale,
		UpdateParameters: true,

		Subject:  "[InteractiveSolutions/Communications] template missing",
		TextBody: fmt.Sprintf("A template is missing for template id: %s, locale: %s", templateId, locale),
		HtmlBody: fmt.Sprintf("A template is missing for template id: %s, locale: %s", templateId, locale),

		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if err := a.templateRepo.Create(&tpl); err != nil {
		return tpl, err
	}

	return tpl, nil
}

func (a *application) getFallbackTemplate(templateId string) (Template, error) {
	tpl, err := a.templateRepo.Get(templateId, a.fallbackLocale)
	switch err {
	case nil:
		return tpl, nil

	case TemplateNotFoundErr:
		return a.createMockTemplate(templateId, a.fallbackLocale)

	default:
		return tpl, err
	}
}

func (a *application) getTemplate(templateId, locale string) (Template, error) {
	tpl, err := a.templateRepo.Get(templateId, locale)
	switch err {
	case nil:
		if tpl.Enabled {
			return tpl, nil
		}

		return a.getFallbackTemplate(templateId)

	case TemplateNotFoundErr:
		if _, err := a.createMockTemplate(templateId, locale); err != nil {
			a.logger.
				WithField("templateId", templateId).
				WithField("locale", locale).
				WithError(err).
				Error("Failed to create mock template")
		}

		return a.getFallbackTemplate(templateId)

	default:
		return tpl, err
	}
}

func (a *application) process(job *Job) error {
	tpl, err := a.getTemplate(job.TemplateId, job.Locale)
	if err != nil {
		return err
	}

	if tpl.UpdateParameters {
		tpl.Parameters = job.Params
		tpl.UpdateParameters = false

		if err := a.templateRepo.Update(&tpl); err != nil {
			return err
		}
	}

	switch job.Type {
	case JobSms:
		return a.defaultSmsTransport.Send(context.Background(), job, tpl, a.render)

	case JobEmail:
		return a.defaultEmailTransport.Send(context.Background(), job, tpl, a.render)

	default:
		return errors.Errorf("Unknown job type %s", job.Type)
	}
}

func (a *application) Render(template Template, job *Job) (subject, text, html string, err error) {
	subject, err = a.render(template.Subject, job.Params)
	if err != nil {
		return
	}

	text, err = a.render(template.TextBody, job.Params)
	if err != nil {
		return
	}

	html, err = a.render(template.HtmlBody, job.Params)
	return
}

func (a *application) render(body string, params map[string]interface{}) (string, error) {
	if params == nil {
		params = map[string]interface{}{}
	}

	for key, value := range a.staticParams {
		if _, ok := params[key]; ok {
			// Allow dynamic parameters to overwrite static parameters
			continue
		}

		params[key] = value
	}

	tpl, err := template.New("").Funcs(a.templateFuncMap).Parse(body)
	if err != nil {
		return "", err
	}

	out := &bytes.Buffer{}

	if err := tpl.Execute(out, params); err != nil {
		return "", err
	}

	return out.String(), nil
}
