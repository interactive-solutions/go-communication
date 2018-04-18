package provider

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type sesTransport struct {
	ses *ses.SES

	from    string
	charset string
}

func NewSesTransport(sess *session.Session, from string) *sesTransport {
	return &sesTransport{
		ses:     ses.New(sess),
		from:    from,
		charset: "UTF-8",
	}
}

func (transport *sesTransport) Send(ctx context.Context, email, subject, textBody, htmlBody string) error {
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(email),
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
	_, err := transport.ses.SendEmail(input)
	if err != nil {
		return err
	}

	return nil
}
