package communication

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const UserAgent = "InteractiveSolutions/GoCommunication-1.0"

type Application interface {
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

func SetTransactionRepo(repo TransactionRepository) AppOption {
	return func(a *application) {
		a.transactionRepo = repo
	}
}

type application struct {
	logger logrus.FieldLogger

	workerCtx    context.Context
	workerCancel context.CancelFunc

	workerQueue chan *Job

	templateRepo    TemplateRepository
	transactionRepo TransactionRepository

	fallbackLocale        string
	defaultSmsTransport   SmsTransport
	defaultEmailTransport EmailTransport
}

func NewApplication(options ...AppOption) (Application, error) {
	app := &application{
		logger:      logrus.NewLogger(),
		workerQueue: make(chan *Job, 1000),
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

	for i := 0; i <= 5; i++ {
		go app.worker(ctx)
	}

	jobs, err := app.transactionRepo.GetPending()
	if err != nil {
		return app, err
	}

	for _, job := range jobs {
		app.queue(&job)
	}

	return app, nil
}

func (a *application) ensureUsableConfiguration() error {
	if a.templateRepo == nil {
		return errors.New("Missing template repository")
	}

	if a.transactionRepo == nil {
		return errors.New("Missing transaction repository")
	}

	return nil
}

func (a *application) SendEmail(id, locale, email string, params map[string]interface{}) error {
	if a.defaultEmailTransport == nil {
		return errors.New("No email transport configured")
	}

	job := &Job{
		Type:       JobEmail,
		TemplateId: id,
		Locale:     locale,
		Target:     email,
		Params:     params,
	}

	if err := a.transactionRepo.Create(job); err != nil {
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
		Type:       JobSms,
		TemplateId: id,
		Locale:     locale,
		Target:     number,
		Params:     params,
	}

	if err := a.transactionRepo.Create(job); err != nil {
		return err
	}

	a.queue(job)
}

func (a *application) Shutdown(ctx context.Context) {
	<-ctx.Done()
	a.workerCancel()
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
				// todo: log
			}

			now := time.Now()

			job.SendAt = &now

			if err := a.transactionRepo.Update(job); err != nil {
				// todo: log
			}
		}
	}
}

func (a *application) createMockTemplate(templateId, locale string) (Template, error) {
	tpl := Template{
		TemplateId:       templateId,
		Locale:           locale,
		UpdateParameters: true,
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
		a.createMockTemplate(templateId, locale)

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
		return err
	}

	htmlBody, err := a.render(tpl.HtmlBody, job.Params)
	if err != nil {
		return err
	}

	textBody, err := a.render(tpl.TextBody, job.Params)
	if err != nil {
		return err
	}

	return a.defaultEmailTransport.Send(context.Background(), job.Target, subject, textBody, htmlBody)
}

func (a *application) renderAndSendSms(job *Job, tpl Template) error {
	message, err := a.render(tpl.TextBody, job.Params)
	if err != nil {
		return err
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
