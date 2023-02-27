package rubygems

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Client struct {
	registryURL string

	cli httpcli.Doer

	// Self-imposed rate-limiter.
	limiter *ratelimit.InstrumentedLimiter
}

func NewClient(urn string, registryURL string, cli httpcli.Doer) *Client {
	return &Client{
		registryURL: registryURL,
		cli:         cli,
		limiter:     ratelimit.DefaultRegistry.Get(urn),
	}
}

func (c *Client) GetPackageContents(ctx context.Context, dep reposource.VersionedPackage) (body io.ReadCloser, url string, err error) {
	url = fmt.Sprintf("%s/gems/%s-%s.gem", strings.TrimSuffix(c.registryURL, "/"), dep.PackageSyntax(), dep.PackageVersion())

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, url, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, url, err
	}
	req.Header.Add("User-Agent", "sourcegraph-rubygems-syncer (sourcegraph.com)")

	body, err = c.do(req)
	if err != nil {
		return nil, url, err
	}
	return body, url, nil
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
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			bs = []byte(errors.Wrap(err, "failed to read body").Error())
		}
		return nil, &Error{path: req.URL.Path, code: resp.StatusCode, message: string(bs)}
	}
	return resp.Body, nil
}
