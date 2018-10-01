package gitlab

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/pkg/metrics"
	"github.com/sourcegraph/sourcegraph/pkg/ratelimit"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"golang.org/x/net/context/ctxhttp"
)

var requestCounter = metrics.NewRequestCounter("gitlab", "Total number of requests sent to the GitLab API.")

// Client is a GitLab API client.
type Client struct {
	baseURL    *url.URL     // base URL of GitLab; e.g., https://gitlab.com or https://gitlab.example.com
	token      string       // a personal access token to authenticate requests, if set
	httpClient *http.Client // the HTTP client to use

	RateLimit *ratelimit.Monitor // the API rate limit monitor

	projCache *rcache.Cache
}

// NewClient creates a new GitLab API client with an optional personal access token to authenticate requests.
//
// The URL must point to the base URL of the GitLab instance. This is https://gitlab.com for GitLab.com and
// http[s]://[gitlab-hostname] for self-hosted GitLab instances.
func NewClient(baseURL *url.URL, token string, transport http.RoundTripper) *Client {
	if transport == nil {
		transport = http.DefaultTransport
	}
	transport = requestCounter.Transport(transport, func(u *url.URL) string {
		// The 3rd component of the Path (/api/v4/XYZ) mostly maps to the type of API
		// request we are making.
		var category string
		if parts := strings.SplitN(u.Path, "/", 3); len(parts) >= 4 {
			category = parts[3]
		}
		return category
	})

	var cacheTTL time.Duration
	if isGitLabDotComURL(baseURL) && token == "" {
		cacheTTL = 10 * time.Minute // cache for longer when unauthenticated
	} else {
		cacheTTL = 30 * time.Second
	}

	// Cache for GitLab project metadata.
	key := sha256.Sum256([]byte(token + ":" + baseURL.String()))
	projCache := rcache.NewWithTTL("gl_proj:"+base64.URLEncoding.EncodeToString(key[:]), int(cacheTTL/time.Second))

	return &Client{
		baseURL:    baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, "api/v4") + "/"}),
		token:      token,
		httpClient: &http.Client{Transport: transport},
		RateLimit:  &ratelimit.Monitor{},
		projCache:  projCache,
	}
}

func isGitLabDotComURL(baseURL *url.URL) bool {
	hostname := strings.ToLower(baseURL.Hostname())
	return hostname == "gitlab.com" || hostname == "www.gitlab.com"
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) (responseHeader http.Header, err error) {
	req.URL = c.baseURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.token != "" {
		req.Header.Set("Private-Token", c.token) // https://docs.gitlab.com/ee/api/README.html#personal-access-tokens
	}

	var resp *http.Response

	span, ctx := opentracing.StartSpanFromContext(ctx, "GitLab")
	span.SetTag("URL", req.URL.String())
	defer func() {
		if err != nil {
			span.SetTag("error", err.Error())
		}
		if resp != nil {
			span.SetTag("status", resp.Status)
		}
		span.Finish()
	}()

	resp, err = ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.RateLimit.Update(resp.Header)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(httpError(resp.StatusCode), fmt.Sprintf("unexpected response from GitLab API (%s)", req.URL))
	}

	return resp.Header, json.NewDecoder(resp.Body).Decode(result)
}

type httpError int

func (err httpError) Error() string {
	return fmt.Sprintf("HTTP error status %d", err)
}

// HTTPErrorCode returns err's HTTP status code, if it is an HTTP error from
// this package. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	e, ok := err.(httpError)
	if !ok {
		// Try one level deeper.
		err = errors.Cause(err)
		e, ok = err.(httpError)
	}
	if ok {
		return int(e)
	}
	return 0
}

// ErrNotFound is when the requested GitLab project is not found.
var ErrNotFound = errors.New("GitLab project not found")

// IsNotFound reports whether err is a GitLab API error of type NOT_FOUND, the equivalent cached
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	if err == ErrNotFound || errors.Cause(err) == ErrNotFound {
		return true
	}
	if HTTPErrorCode(err) == http.StatusNotFound {
		return true
	}
	return false
}
