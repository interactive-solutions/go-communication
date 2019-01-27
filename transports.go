package communication

import "context"

type Transport interface {
	Send(ctx context.Context, job *Job, template Template, render RenderFunc) error
}

