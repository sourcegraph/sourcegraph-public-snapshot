//nolint:bodyclose // Body is closed in Client.Do, but the response is still returned to provide access to the headers
package gerrit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	defaultRateLimit      = rate.Limit(8) // 480/min or 28,800/hr
	defaultRateLimitBurst = 500
)

// Client access a Gerrit via the REST API.
type Client struct {
	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// Config is the code host connection config for this client
	Config *schema.GerritConnection

	// URL is the base URL of Gerrit.
	URL *url.URL

	// RateLimit is the self-imposed rate limiter.
	RateLimit *rate.Limiter

	// NoAuth determines whether the calls made by the client should use authentication.
	NoAuth bool
}

// NewClient returns an authenticated Gerrit API client with
// the provided configuration. If a nil httpClient is provided, http.DefaultClient
// will be used.
func NewClient(config *schema.GerritConnection, httpClient httpcli.Doer) (*Client, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	// Normally our registry will return a default infinite limiter when nothing has been
	// synced from config. However, we always want to ensure there is at least some form of rate
	// limiting for Gerrit.
	defaultLimiter := rate.NewLimiter(defaultRateLimit, defaultRateLimitBurst)
	l := ratelimit.DefaultRegistry.GetOrSet(u.String(), defaultLimiter)

	return &Client{
		httpClient: httpClient,
		Config:     config,
		URL:        u,
		RateLimit:  l,
		NoAuth:     false,
	}, nil
}

// ListProjectsArgs defines options to be set on ListProjects method calls.
type ListProjectsArgs struct {
	Cursor *Pagination
}

// ListProjectsResponse defines a response struct returned from ListProjects method calls.
type ListProjectsResponse map[string]*Project

// SetNoAuth used for testing
func (c *Client) SetNoAuth(auth bool) {
	c.NoAuth = auth
}

func (c *Client) ListProjects(ctx context.Context, opts ListProjectsArgs) (*ListProjectsResponse, error) {
	qs := make(url.Values)

	if opts.Cursor == nil {
		opts.Cursor = &Pagination{PerPage: 100, Page: 1}
	}

	qs.Set("type", "CODE")
	qs.Set("n", fmt.Sprintf("%d", opts.Cursor.PerPage))
	qs.Set("S", fmt.Sprintf("%d", (opts.Cursor.Page-1)*opts.Cursor.PerPage))

	urlPath := "projects/"
	// Add a prefix for authenticated requests.
	if !c.NoAuth {
		urlPath = "a/" + urlPath
	}

	u := url.URL{Path: urlPath, RawQuery: qs.Encode()}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	var resp ListProjectsResponse
	if _, err = c.do(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) (*http.Response, error) {
	req.URL = c.URL.ResolveReference(req.URL)

	// Add Basic Auth headers for authenticated requests.
	if !c.NoAuth {
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.Config.Username+":"+c.Config.Password)))
	}

	startWait := time.Now()
	if err := c.RateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	if d := time.Since(startWait); d > 200*time.Millisecond {
		log15.Warn("Gerrit self-enforced API rate limit: request delayed longer than expected due to rate limit", "delay", d)
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, errors.WithStack(&httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		})
	}

	// The first 4 characters of the Gerrit API responses need to be stripped, see: https://gerrit-review.googlesource.com/Documentation/rest-api.html#output .
	return resp, json.Unmarshal(bs[4:], result)
}

type Project struct {
	Description string            `json:"description"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Parent      string            `json:"parent"`
	State       string            `json:"state"`
	Branches    map[string]string `json:"branches"`
	Labels      map[string]Label  `json:"labels"`
}

type Label struct {
	Values       map[string]string `json:"values"`
	DefaultValue string            `json:"default_value"`
}

type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Gerrit API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
