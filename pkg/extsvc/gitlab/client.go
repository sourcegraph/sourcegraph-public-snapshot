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
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/metrics"
	"github.com/sourcegraph/sourcegraph/pkg/ratelimit"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"golang.org/x/net/context/ctxhttp"
)

var requestCounter = metrics.NewRequestCounter("gitlab", "Total number of requests sent to the GitLab API.")

// ClientProvider creates GitLab API clients. Each client has separate authentication creds and a
// separate cache, but they share an underlying HTTP client and rate limiter. Callers who want a simple
// unauthenticated API client should use `NewClientProvider(baseURL, transport).GetClient()`.
type ClientProvider struct {
	// baseURL is the base URL of GitLab; e.g., https://gitlab.com or https://gitlab.example.com
	baseURL *url.URL

	// httpClient is the underlying the HTTP client to use
	httpClient *http.Client

	gitlabClients   map[string]*Client
	gitlabClientsMu sync.Mutex

	RateLimit *ratelimit.Monitor // the API rate limit monitor
}

type CommonOp struct {
	NoCache bool
}

func NewClientProvider(baseURL *url.URL, transport http.RoundTripper) *ClientProvider {
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

	return &ClientProvider{
		baseURL:       baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, "api/v4") + "/"}),
		httpClient:    &http.Client{Transport: transport},
		gitlabClients: make(map[string]*Client),
		RateLimit:     &ratelimit.Monitor{},
	}
}

// GetPATClient returns a client authenticated by the personal access token.
func (p *ClientProvider) GetPATClient(personalAccessToken string) *Client {
	if personalAccessToken == "" {
		return p.getClient("", "", "")
	}
	return p.getClient(fmt.Sprintf("pat::%s", personalAccessToken), personalAccessToken, "")
}

// GetOAuthClient returns a client authenticated by the OAuth token.
func (p *ClientProvider) GetOAuthClient(oauthToken string) *Client {
	if oauthToken == "" {
		return p.getClient("", "", "")
	}
	return p.getClient(fmt.Sprintf("oauth::%s", oauthToken), "", oauthToken)
}

// GetClient returns an unauthenticated client.
func (p *ClientProvider) GetClient() *Client {
	return p.getClient("", "", "")
}

func (p *ClientProvider) getClient(key, personalAccessToken, oauthToken string) *Client {
	p.gitlabClientsMu.Lock()
	defer p.gitlabClientsMu.Unlock()

	if c, ok := p.gitlabClients[key]; ok {
		return c
	}

	c := p.newClient(p.baseURL, personalAccessToken, oauthToken, p.httpClient, p.RateLimit)
	p.gitlabClients[key] = c
	return c
}

type Client struct {
	baseURL             *url.URL
	httpClient          *http.Client
	projCache           *rcache.Cache
	personalAccessToken string // a personal access token to authenticate requests, if set
	OAuthToken          string // an OAuth bearer token, if set
	RateLimit           *ratelimit.Monitor
}

// newClient creates a new GitLab API client with an optional personal access token to authenticate requests.
//
// The URL must point to the base URL of the GitLab instance. This is https://gitlab.com for GitLab.com and
// http[s]://[gitlab-hostname] for self-hosted GitLab instances.
func (p *ClientProvider) newClient(baseURL *url.URL, personalAccessToken, oauthToken string, httpClient *http.Client, rateLimit *ratelimit.Monitor) *Client {
	// Cache for GitLab project metadata.
	var cacheTTL time.Duration
	if isGitLabDotComURL(baseURL) && personalAccessToken == "" && oauthToken == "" {
		cacheTTL = 10 * time.Minute // cache for longer when unauthenticated
	} else {
		cacheTTL = 30 * time.Second
	}
	key := sha256.Sum256([]byte(personalAccessToken + ":" + oauthToken + ":" + baseURL.String()))
	projCache := rcache.NewWithTTL("gl_proj:"+base64.URLEncoding.EncodeToString(key[:]), int(cacheTTL/time.Second))

	return &Client{
		baseURL:             baseURL,
		httpClient:          httpClient,
		projCache:           projCache,
		personalAccessToken: personalAccessToken,
		OAuthToken:          oauthToken,
		RateLimit:           rateLimit,
	}
}

func isGitLabDotComURL(baseURL *url.URL) bool {
	hostname := strings.ToLower(baseURL.Hostname())
	return hostname == "gitlab.com" || hostname == "www.gitlab.com"
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) (responseHeader http.Header, err error) {
	req.URL = c.baseURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.personalAccessToken != "" {
		req.Header.Set("Private-Token", c.personalAccessToken) // https://docs.gitlab.com/ee/api/README.html#personal-access-tokens
	}
	if c.OAuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.OAuthToken))
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
