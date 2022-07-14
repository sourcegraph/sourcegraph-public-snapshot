package githubwebhook

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	httpClient httpcli.Doer
}

func NewClient(httpClient httpcli.Doer) (*Client, error) {
	return &Client{
		httpClient: httpClient,
	}, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Newf("non-2XX status code:", resp.StatusCode)
	}

	return resp, nil
}
