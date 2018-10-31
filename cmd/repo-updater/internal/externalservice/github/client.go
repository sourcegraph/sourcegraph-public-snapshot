package github

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/metrics"
	"github.com/sourcegraph/sourcegraph/pkg/ratelimit"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"golang.org/x/net/context/ctxhttp"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degredation"))
	githubProxyURL   = func() *url.URL {
		url, err := url.Parse(env.Get("GITHUB_BASE_URL", "http://github-proxy", "base URL for GitHub.com API (used for github-proxy)"))
		if err != nil {
			log.Fatal("Error parsing GITHUB_BASE_URL:", err)
		}
		return url
	}()

	requestCounter = metrics.NewRequestCounter("github", "Total number of requests sent to the GitHub API.")
)

// Client is a GitHub API client.
type Client struct {
	apiURL       *url.URL     // base URL of GitHub API; e.g., https://api.github.com
	githubDotCom bool         // true if this client connects to github.com
	token        string       // a personal access token to authenticate requests, if set
	httpClient   *http.Client // the HTTP client to use

	repoCache *rcache.Cache

	RateLimit *ratelimit.Monitor // the API rate limit monitor
}

type githubAPIError struct {
	URL              string
	Code             int
	Message          string
	DocumentationURL string `json:"documentation_url"`
}

func (e *githubAPIError) Error() string {
	return fmt.Sprintf("request to %s returned status %d: %s", e.URL, e.Code, e.Message)
}

func urlIsGitHubDotCom(apiURL *url.URL) bool {
	hostname := strings.ToLower(apiURL.Hostname())
	return hostname == "api.github.com" || hostname == "github.com" || hostname == "www.github.com" || apiURL.String() == githubProxyURL.String()
}

func canonicalizedURL(apiURL *url.URL) *url.URL {
	if urlIsGitHubDotCom(apiURL) {
		// For GitHub.com API requests, use github-proxy (which adds our OAuth2 client ID/secret to get a much higher
		// rate limit).
		return githubProxyURL
	}
	return apiURL
}

// NewRepoCache creates a new cache for GitHub repository metadata.
func NewRepoCache(apiURL *url.URL, token string) *rcache.Cache {
	apiURL = canonicalizedURL(apiURL)
	var cacheTTL time.Duration
	if urlIsGitHubDotCom(apiURL) {
		cacheTTL = 10 * time.Minute
	} else {
		// GitHub Enterprise
		cacheTTL = 30 * time.Second
	}
	key := sha256.Sum256([]byte(token + ":" + apiURL.String()))
	return rcache.NewWithTTL("gh_repo:"+base64.URLEncoding.EncodeToString(key[:]), int(cacheTTL/time.Second))
}

// NewClient creates a new GitHub API client with an optional personal access token to authenticate
// requests.
//
// The API URL must point to the base URL of the GitHub API. This is https://api.github.com for
// GitHub.com and http[s]://[github-enterprise-hostname]/api for GitHub Enterprise.
// repoCache should be an existing repository cache created with
// NewRepoCache, or nil to create a new one. Clients querying the same
// host should use the same cache (see repos.githubConnection, which uses
// multiple clients to track independent rate limits in the same host).
func NewClient(apiURL *url.URL, token string, transport http.RoundTripper, repoCache *rcache.Cache) *Client {
	apiURL = canonicalizedURL(apiURL)
	if gitHubDisable {
		transport = disabledTransport{}
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	transport = requestCounter.Transport(transport, func(u *url.URL) string {
		// The first component of the Path mostly maps to the type of API
		// request we are making. See `curl https://api.github.com` for the
		// exact mapping
		var category string
		if parts := strings.SplitN(u.Path, "/", 3); len(parts) > 1 {
			category = parts[1]
		}
		return category
	})

	if repoCache == nil {
		// Create a new cache for repository metadata.
		repoCache = NewRepoCache(apiURL, token)
	}

	return &Client{
		apiURL:       apiURL,
		githubDotCom: urlIsGitHubDotCom(apiURL),
		token:        token,
		httpClient:   &http.Client{Transport: transport},
		RateLimit:    &ratelimit.Monitor{HeaderPrefix: "X-"},
		repoCache:    repoCache,
	}
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) (err error) {
	req.URL.Path = path.Join(c.apiURL.Path, req.URL.Path)
	req.URL = c.apiURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.token != "" {
		req.Header.Set("Authorization", "bearer "+c.token)
	}

	var resp *http.Response

	span, ctx := opentracing.StartSpanFromContext(ctx, "GitHub")
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
		return err
	}
	defer resp.Body.Close()
	c.RateLimit.Update(resp.Header)
	if resp.StatusCode != http.StatusOK {
		var err githubAPIError
		if decErr := json.NewDecoder(resp.Body).Decode(&err); decErr != nil {
			log15.Warn("Failed to decode error response from github API", "error", decErr)
		}
		err.URL = req.URL.String()
		err.Code = resp.StatusCode
		return &err
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) requestGet(ctx context.Context, requestURI string, result interface{}) error {
	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		return err
	}

	// Include node_id (GraphQL ID) in response. See
	// https://developer.github.com/changes/2017-12-19-graphql-node-id/.
	req.Header.Add("Accept", "application/vnd.github.jean-grey-preview+json")

	return c.do(ctx, req, result)
}

func (c *Client) requestGraphQL(ctx context.Context, query string, vars map[string]interface{}, result interface{}) (err error) {
	reqBody, err := json.Marshal(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", "/graphql", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	var respBody struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}
	if err := c.do(ctx, req, &respBody); err != nil {
		return err
	}
	if len(respBody.Errors) > 0 {
		return respBody.Errors
	}
	if result != nil && respBody.Data != nil {
		if err := unmarshal(respBody.Data, result); err != nil {
			return err
		}
	}
	return nil
}

// unmarshal wraps json.Unmarshal, but includes extra context in the case of
// json.UnmarshalTypeError
func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if e, ok := err.(*json.UnmarshalTypeError); ok && e.Offset >= 0 {
		a := e.Offset - 100
		b := e.Offset + 100
		if a < 0 {
			a = 0
		}
		if b > int64(len(data)) {
			b = int64(len(data))
		}
		if e.Offset >= int64(len(data)) {
			return errors.Wrapf(err, "graphql: cannot unmarshal at offset %d: before %q", e.Offset, string(data[a:e.Offset]))
		}
		return errors.Wrapf(err, "graphql: cannot unmarshal at offset %d: before %q; after %q", e.Offset, string(data[a:e.Offset]), string(data[e.Offset:b]))
	}
	return err
}

// HTTPErrorCode returns err's HTTP status code, if it is an HTTP error from
// this package. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	if e, ok := errors.Cause(err).(*githubAPIError); ok {
		return e.Code
	}
	return 0
}

// ErrNotFound is when the requested GitHub repository is not found.
var ErrNotFound = errors.New("GitHub repository not found")

// IsNotFound reports whether err is a GitHub API error of type NOT_FOUND, the equivalent cached
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	if err == ErrNotFound || errors.Cause(err) == ErrNotFound {
		return true
	}
	if HTTPErrorCode(err) == http.StatusNotFound {
		return true
	}
	errs, ok := err.(graphqlErrors)
	if !ok {
		return false
	}
	for _, err := range errs {
		if err.Type == "NOT_FOUND" {
			return true
		}
	}
	return false
}

// IsRateLimitExceeded reports whether err is a GitHub API error reporting that the GitHub API rate
// limit was exceeded.
func IsRateLimitExceeded(err error) bool {
	if e, ok := errors.Cause(err).(*githubAPIError); ok {
		return strings.Contains(e.Message, "API rate limit exceeded") || strings.Contains(e.DocumentationURL, "#rate-limiting")
	}

	errs, ok := err.(graphqlErrors)
	if !ok {
		return false
	}
	for _, err := range errs {
		// This error is not documented, so be lenient here (instead of just checking for exact
		// error type match.)
		if err.Type == "RATE_LIMITED" || strings.Contains(err.Message, "API rate limit exceeded") {
			return true
		}
	}
	return false
}

// graphqlErrors describes the errors in a GraphQL response. It contains at least 1 element when returned by
// requestGraphQL. See https://facebook.github.io/graphql/#sec-Errors.
type graphqlErrors []struct {
	Message   string   `json:"message"`
	Type      string   `json:"type"`
	Path      []string `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

func (e graphqlErrors) Error() string {
	return fmt.Sprintf("error in GraphQL response: %s", e[0].Message)
}

type disabledTransport struct{}

func (t disabledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communication disabled")
}
