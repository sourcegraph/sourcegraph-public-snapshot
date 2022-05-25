package crates

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	// A list of crates proxies.
	urls []string

	cli httpcli.Doer

	// Self-imposed rate-limiter.
	limiter *ratelimit.InstrumentedLimiter
}

type RustFile struct {
	Name string
	URL  string
}

func NewClient(urn string, urls []string, cli httpcli.Doer) *Client {
	return &Client{
		urls:    urls,
		cli:     cli,
		limiter: ratelimit.DefaultRegistry.Get(urn),
	}
}

func (c *Client) Version(ctx context.Context, name string, version string) (*RustFile, error) {
	return &RustFile{
		Name: name,
		URL:  fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s/download", name, version),
	}, nil
}

func (c *Client) Download(ctx context.Context, url string) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	b, err := c.do(req)
	if err != nil {
		return nil, errors.Wrap(err, "crates")
	}
	return b, nil
}

// TODO: This is just copied from the python one? {{{
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

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{path: req.URL.Path, code: resp.StatusCode, message: string(bs)}
	}

	return bs, nil
}

// }}}
