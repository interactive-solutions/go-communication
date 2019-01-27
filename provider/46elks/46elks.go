package elks

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/interactive-solutions/go-communication"
	"github.com/pkg/errors"
)

const elksApi = "https://api.46elks.com/a1/sms"

// Elks in an implementation for 46elks
type elks struct {
	client *retryablehttp.Client

	from string

	username string
	password string
}

func New46ElksClient(from, username, password string) communication.Transport {
	return &elks{
		client: retryablehttp.NewClient(),

		from:     from,
		username: username,
		password: password,
	}
}

func (e *elks) Send(ctx context.Context, job *communication.Job, template communication.Template, render communication.RenderFunc) error {
	message, err := render(template.TextBody, job.Params)
	if err != nil {
		return errors.Wrap(err, "Failed to generate sms message from template")
	}

	body := url.Values{
		"from":    {e.from},
		"to":      {job.Target},
		"message": {message},
	}.Encode()

	req, err := retryablehttp.NewRequest(http.MethodPost, elksApi, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.SetBasicAuth(e.username, e.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	req.Header.Set("User-Agent", communication.UserAgent)

	if resp, err := e.client.Do(req); err != nil {
		return err
	} else if resp.StatusCode >= 300 || resp.StatusCode <= 199 {
		return errors.Errorf("Unexpected response code %d received from 46elks", resp.StatusCode)
	}

	return nil
}
