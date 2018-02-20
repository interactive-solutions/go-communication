package mailgun

import (
	"context"

	"github.com/interactive-solutions/go-communication"
	"gopkg.in/mailgun/mailgun-go.v1"
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

func NewMailgunTransport(mailgunClient mailgun.Mailgun, options ...MailgunOption) communication.EmailTransport {
	t := &mailgunTransport{
		mg: mailgunClient,
	}

	for _, option := range options {
		option(t)
	}

	return t
}

func (t *mailgunTransport) Send(ctx context.Context, email, subject, textBody, htmlBody string) error {

	msg := t.mg.NewMessage(t.from, subject, textBody, email)
	msg.SetHtml(htmlBody)

	if t.replyTo != "" {
		msg.SetReplyTo(t.replyTo)
	}

	_, _, err := t.mg.Send(msg)
	return err
}
