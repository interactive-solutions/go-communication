package provider

import (
	"context"
	"github.com/interactive-solutions/go-communication"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type sesTransport struct {
	ses *ses.SES

	from    string
	charset string
}

func NewSesTransport(sess *session.Session, from string) communication.Transport {
	return &sesTransport{
		ses:     ses.New(sess),
		from:    from,
		charset: "UTF-8",
	}
}

func (transport *sesTransport) Send(ctx context.Context, job *communication.Job, template communication.Template, render communication.RenderFunc) error {
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

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(job.Target),
			},
		},
		Tags: []*ses.MessageTag{
			{
				Name:  aws.String("template"),
				Value: aws.String(template.TemplateId),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(transport.charset),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(transport.charset),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(transport.charset),
				Data:    aws.String(subject),
			},
		},

		Source: aws.String(transport.from),
	}

	// Attempt to send the email.
	_, err = transport.ses.SendEmail(input)
	return errors.Wrap(err, "Failed to send email")
}
