package crates

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

type Client struct {
	cli httpcli.Doer

	// Self-imposed rate-limiter.
	limiter *ratelimit.InstrumentedLimiter
}

func NewClient(urn string, cli httpcli.Doer) *Client {
	return &Client{
		cli:     cli,
		limiter: ratelimit.DefaultRegistry.Get(urn),
	}
}

func (c *Client) Get(ctx context.Context, url string) (io.ReadCloser, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "sourcegraph-crates-syncer (sourcegraph.com)")

	b, err := c.do(req)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type Error struct {
	path    string
	code    int
	message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bad response with status code %d for %s: %s", e.code, e.path, e.message)
}

func (e *Error) NotFound() bool {
	return e.code == http.StatusNotFound
}

func (c *Client) do(req *http.Request) (io.ReadCloser, error) {
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, &Error{path: req.URL.Path, code: resp.StatusCode, message: fmt.Sprintf("failed to read non-200 body: %v", err)}
		}
		return nil, &Error{path: req.URL.Path, code: resp.StatusCode, message: string(bs)}
	}

	return resp.Body, nil
}
