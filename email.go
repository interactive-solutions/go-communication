package communication

import "context"

type EmailTransport interface {
	Send(ctx context.Context, email, subject, textBody, htmlBody string) error
}
