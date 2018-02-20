package communication

import "context"

type SmsTransport interface {
	Send(ctx context.Context, number string, message string) error
}
