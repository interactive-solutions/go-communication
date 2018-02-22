package communication

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const UserAgent = "InteractiveSolutions/GoCommunication-1.0"

type Application interface {
	HttpHandler() *httpHandler
	SendEmail(id, locale, email string, params map[string]interface{}) error
	SendSms(id, locale, number string, params map[string]interface{}) error
	Shutdown(ctx context.Context)
}

type AppOption func(a *application)

func SetFallbackLocale(locale string) AppOption {
	return func(a *application) {
		a.fallbackLocale = locale
	}
}

func SetDefaultSmsTransport(transport SmsTransport) AppOption {
	return func(a *application) {
		a.defaultSmsTransport = transport
	}
}

func SetDefaultEmailTransport(transport EmailTransport) AppOption {
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

type application struct {
	logger logrus.FieldLogger

	workerCtx    context.Context
	workerCancel context.CancelFunc

	workerQueue chan *Job
	workerCount int

	templateRepo TemplateRepository
	jobRepo      JobRepository

	fallbackLocale        string
	defaultSmsTransport   SmsTransport
	defaultEmailTransport EmailTransport
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
		app.queue(&job)
	}

	return app, nil
}

func (a *application) HttpHandler() *httpHandler {
	return &httpHandler{
		app: a,
	}
}

func (a *application) SendEmail(id, locale, email string, params map[string]interface{}) error {
	if a.defaultEmailTransport == nil {
		return errors.New("No email transport configured")
	}

	job := &Job{
		Uuid:       uuid.NewV4(),
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

func (a *application) SendSms(id, locale, number string, params map[string]interface{}) error {
	if a.defaultSmsTransport == nil {
		return errors.New("No sms transport configured")
	}

	job := &Job{
		Uuid:       uuid.NewV4(),
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
		return a.renderAndSendSms(job, tpl)

	case JobEmail:
		return a.renderAndSendEmail(job, tpl)

	default:
		return errors.Errorf("Unknown job type %d", job.Type)
	}
}

func (a *application) renderAndSendEmail(job *Job, tpl Template) error {

	subject, err := a.render(tpl.Subject, job.Params)
	if err != nil {
		return errors.Wrap(err, "failed to parse subject")
	}

	htmlBody, err := a.render(tpl.HtmlBody, job.Params)
	if err != nil {
		return errors.Wrap(err, "failed to parse html body")
	}

	textBody, err := a.render(tpl.TextBody, job.Params)
	if err != nil {
		return errors.Wrap(err, "failed to parse text body")
	}

	return a.defaultEmailTransport.Send(context.Background(), job.Target, subject, textBody, htmlBody)
}

func (a *application) renderAndSendSms(job *Job, tpl Template) error {
	message, err := a.render(tpl.TextBody, job.Params)
	if err != nil {
		return errors.Wrap(err, "failed to parse text body")
	}

	return a.defaultSmsTransport.Send(context.Background(), job.Target, message)
}

func (a *application) render(body string, params map[string]interface{}) (string, error) {
	tpl, err := template.New("").Parse(body)
	if err != nil {
		return "", err
	}

	out := &bytes.Buffer{}

	if err := tpl.Execute(out, params); err != nil {
		return "", err
	}

	return out.String(), nil
}
