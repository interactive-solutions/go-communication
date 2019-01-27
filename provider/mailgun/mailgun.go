package mailgun

import (
	"context"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/pkg/errors"

	"github.com/interactive-solutions/go-communication"
)

type MailgunOption func(t *mailgunTransport) error

func SetFrom(from string) MailgunOption {
	return func(e *mailgunTransport) error {
		e.from = from
		return nil
	}
}

func SetReplyTo(reployTo string) MailgunOption {
	return func(e *mailgunTransport) error {
		e.replyTo = reployTo
		return nil
	}
}

type mailgunTransport struct {
	mg mailgun.Mailgun

	from    string
	replyTo string
}

func NewMailgunTransport(mailgunClient mailgun.Mailgun, options ...MailgunOption) communication.Transport {
	t := &mailgunTransport{
		mg: mailgunClient,
	}

	for _, option := range options {
		option(t)
	}

	return t
}

func (t *mailgunTransport) Send(ctx context.Context, job *communication.Job, template communication.Template, render communication.RenderFunc) error {

	subject, err := render(template.Subject, job.Params)
	if err != nil {
		return errors.Wrapf(err, "Failed to render subject for job %s template %s", job.Uuid, template.TemplateId)
	}

	textBody, err := render(template.TextBody, job.Params)
	if err != nil {
		return errors.Wrapf(err, "Failed to render text body for job %s template %s", job.Uuid, template.TemplateId)
	}

	htmlBody, err := render(template.HtmlBody, job.Params)
	if err != nil {
		return errors.Wrapf(err, "Failed to render html body for job %s template %s", job.Uuid, template.TemplateId)
	}

	msg := t.mg.NewMessage(t.from, subject, textBody, job.Target)
	msg.SetHtml(htmlBody)

	if err := msg.AddTag(template.TemplateId, template.Locale); err != nil {
		return errors.Wrap(err, "Failed to add tags")
	}

	if t.replyTo != "" {
		msg.SetReplyTo(t.replyTo)
	}

	_, _, err = t.mg.Send(ctx, msg)
	return errors.Wrap(err, "Failed to send message")
}
