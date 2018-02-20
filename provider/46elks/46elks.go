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

type ElkOption func(elk *elks) error

func SetBasicAuth(username, password string) ElkOption {
	return func(e *elks) error {
		e.username = username
		e.password = password

		return nil
	}
}

func SetFrom(from string) ElkOption {
	return func(e *elks) error {
		e.from = from

		return nil
	}
}

// Elks in an implementation for 46elks
type elks struct {
	client *retryablehttp.Client

	from string

	username string
	password string
}

func New46ElksClient(options ...ElkOption) (*elks, error) {
	client := &elks{
		client: retryablehttp.NewClient(),
	}

	for _, option := range options {
		if err := option(client); err != nil {
			return client, err
		}
	}

	if client.username == "" || client.password == "" {
		return client, errors.New("Missing username/password")
	}

	if client.from == "" {
		return client, errors.New("missing from number/name")
	}

	return client, nil
}

func (e *elks) Send(ctx context.Context, number string, message string) error {

	body := url.Values{
		"from":    {e.from},
		"to":      {number},
		"message": {message},
	}.Encode()

	req, err := retryablehttp.NewRequest(http.MethodPost, elksApi, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}

	req.Request = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	req.Header.Set("User-Agent", communication.UserAgent)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 || resp.StatusCode <= 199 {
		return errors.Errorf("Unexpected response code %d received from 46elks", resp.StatusCode)
	}

	return nil
}
