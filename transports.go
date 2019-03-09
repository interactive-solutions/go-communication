package communication

import "context"

type Transport interface {
	Send(ctx context.Context, job *Job, template Template, render RenderFunc) error
}

type TransportSupportsSubscriptionBlocking interface {
	GetUnsubscribedTemplates(ctx context.Context, email string) ([]string, error)

	ResubscribeToAll(ctx context.Context, email string) error
	ResubscribeToTemplate(ctx context.Context, email, template string) error
}