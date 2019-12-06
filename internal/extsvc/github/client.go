package github

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
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

	requestCounter = metrics.NewRequestMeter("github", "Total number of requests sent to the GitHub API.")
)

// Client is a caching GitHub API client.
//
// All instances use a map of rcache.Cache instances for caching (see the `repoCache` field). These
// separate instances have consistent naming prefixes so that different instances will share the
// same Redis cache entries (provided they were computed with the same API URL and access
// token). The cache keys are agnostic of the http.RoundTripper transport.
type Client struct {
	// apiURL is the base URL of a GitHub API. It must point to the base URL of the GitHub API. This
	// is https://api.github.com for GitHub.com and http[s]://[github-enterprise-hostname]/api for
	// GitHub Enterprise.
	apiURL *url.URL

	// githubDotCom is true if this client connects to github.com.
	githubDotCom bool

	// defaultToken is the personal access token used to authenticate requests if none is specified
	// explicitly in a method call. May be empty, in which case the default behavior is to make
	// unauthenticated requests.
	defaultToken string

	// httpClient is the HTTP client used to make requests to the GitHub API.
	httpClient httpcli.Doer

	// repoCache is a map of rcache.Cache instances keyed by auth token.
	repoCache   map[string]*rcache.Cache
	repoCacheMu sync.RWMutex

	// repoCachePrefix is the cache key prefix used when constructing a new rcache.Cache instance
	// for repoCache. It should normally be left empty. It is only used in tests. Specifying a
	// non-empty value will result in different Redis key values.
	repoCachePrefix string

	// repoCacheTTL is the TTL of cache entries.
	repoCacheTTL time.Duration

	// RateLimit is the API rate limit monitor.
	RateLimit *ratelimit.Monitor
}

// APIError is an error type returned by Client when the GitHub API responds with
// an error.
type APIError struct {
	URL              string
	Code             int
	Message          string
	DocumentationURL string `json:"documentation_url"`
}

func (e *APIError) Error() string {
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

// NewRepoCache creates a new cache for GitHub repository metadata. The backing store is Redis. A
// checksum of the access token and API URL are used as a Redis key prefix to prevent collisions
// with caches for different tokens and API URLs. An optional keyPrefix may also be specified,
// typically used in tests.
func NewRepoCache(apiURL *url.URL, token, keyPrefix string, cacheTTL time.Duration) *rcache.Cache {
	if keyPrefix == "" {
		keyPrefix = "gh_repo:"
	}

	apiURL = canonicalizedURL(apiURL)
	if cacheTTL == 0 {
		if urlIsGitHubDotCom(apiURL) {
			cacheTTL = 10 * time.Minute
		} else {
			// GitHub Enterprise
			cacheTTL = 30 * time.Second
		}
	}
	key := sha256.Sum256([]byte(token + ":" + apiURL.String()))
	return rcache.NewWithTTL(keyPrefix+base64.URLEncoding.EncodeToString(key[:]), int(cacheTTL/time.Second))
}

// NewClient creates a new GitHub API client with an optional default personal access token.
//
// apiURL must point to the base URL of the GitHub API. See the docstring for Client.apiURL.
func NewClient(apiURL *url.URL, defaultToken string, cli httpcli.Doer) *Client {
	apiURL = canonicalizedURL(apiURL)
	if gitHubDisable {
		cli = disabledClient{}
	}
	if cli == nil {
		cli = http.DefaultClient
	}

	cli = requestCounter.Doer(cli, func(u *url.URL) string {
		// The first component of the Path mostly maps to the type of API
		// request we are making. See `curl https://api.github.com` for the
		// exact mapping
		var category string
		if parts := strings.SplitN(u.Path, "/", 3); len(parts) > 1 {
			category = parts[1]
		}
		return category
	})

	return &Client{
		apiURL:       apiURL,
		githubDotCom: urlIsGitHubDotCom(apiURL),
		defaultToken: defaultToken,
		httpClient:   cli,
		RateLimit:    &ratelimit.Monitor{HeaderPrefix: "X-"},
		repoCache:    map[string]*rcache.Cache{},
	}
}

// cache returns the cache associated with the token (which can be empty, in which case the default
// token will be used). Accessors of the caches should use this method rather than referencing
// repoCache directly.
func (c *Client) cache(explicitToken string) *rcache.Cache {
	token := firstNonEmpty(explicitToken, c.defaultToken)

	c.repoCacheMu.RLock()
	if cache, ok := c.repoCache[token]; ok {
		c.repoCacheMu.RUnlock()
		return cache
	}
	c.repoCacheMu.RUnlock()

	c.repoCacheMu.Lock()
	// Recheck that the cache item exists once we acquire the write-lock in case it has been
	// populated.
	defer c.repoCacheMu.Unlock()
	if cache, ok := c.repoCache[token]; ok {
		return cache
	}
	// BUG?
	c.repoCache[token] = NewRepoCache(c.apiURL, token, c.repoCachePrefix, c.repoCacheTTL)
	return c.repoCache[token]
}

func (c *Client) do(ctx context.Context, token string, req *http.Request, result interface{}) (err error) {
	req.URL.Path = path.Join(c.apiURL.Path, req.URL.Path)
	req.URL = c.apiURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "bearer "+token)
	} else if c.defaultToken != "" {
		req.Header.Set("Authorization", "bearer "+c.defaultToken)
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

	resp, err = c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	c.RateLimit.Update(resp.Header)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var err APIError
		if body, readErr := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<13)); readErr != nil { // 8kb
			err.Message = fmt.Sprintf("failed to read error response from GitHub API: %v: %q", readErr, string(body))
		} else if decErr := json.Unmarshal(body, &err); decErr != nil {
			err.Message = fmt.Sprintf("failed to decode error response from GitHub API: %v: %q", decErr, string(body))
		}
		err.URL = req.URL.String()
		err.Code = resp.StatusCode
		return &err
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// listRepositories is a generic method that unmarshals the given
// JSON HTTP endpoint into a []restRepository. It will return an
// error if it fails.
//
// This is used to extract repositories from the GitHub API endpoints:
// - /users/:user/repos
// - /orgs/:org/repos
// - /user/repos
func (c *Client) listRepositories(ctx context.Context, requestURI string) ([]*Repository, error) {
	var restRepos []restRepository
	if err := c.requestGet(ctx, "", requestURI, &restRepos); err != nil {
		return nil, err
	}
	repos := make([]*Repository, 0, len(restRepos))
	for _, restRepo := range restRepos {
		repos = append(repos, convertRestRepo(restRepo))
	}
	return repos, nil
}

// ListInstallationRepositories lists repositories on which the authenticated
// GitHub App has been installed.
func (c *Client) ListInstallationRepositories(ctx context.Context) ([]*Repository, error) {
	type response struct {
		Repositories []restRepository `json:"repositories"`
	}
	var resp response
	if err := c.requestGet(ctx, "", "installation/repositories", &resp); err != nil {
		return nil, err
	}
	repos := make([]*Repository, 0, len(resp.Repositories))
	for _, restRepo := range resp.Repositories {
		repos = append(repos, convertRestRepo(restRepo))
	}
	return repos, nil
}

func (c *Client) requestGet(ctx context.Context, token, requestURI string, result interface{}) error {
	req, err := http.NewRequest("GET", requestURI, nil)
	if err != nil {
		return err
	}

	// Include node_id (GraphQL ID) in response. See
	// https://developer.github.com/changes/2017-12-19-graphql-node-id/.
	//
	// Enable the repository topics API. See
	// https://developer.github.com/v3/repos/#list-all-topics-for-a-repository
	req.Header.Add("Accept", "application/vnd.github.jean-grey-preview+json,application/vnd.github.mercy-preview+json")

	// Enable the GitHub App API. See
	// https://developer.github.com/v3/apps/installations/#list-repositories
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	return c.do(ctx, token, req, result)
}

func (c *Client) requestGraphQL(ctx context.Context, token, query string, vars map[string]interface{}, result interface{}) (err error) {
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

	// GitHub.com GraphQL endpoint is api.github.com/graphql. GitHub Enterprise is /api/graphql (the
	// REST endpoint is /api/v3, necessitating the "..").
	graphqlEndpoint := "/graphql"
	if !c.githubDotCom {
		graphqlEndpoint = "../graphql"
	}
	req, err := http.NewRequest("POST", graphqlEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	var respBody struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}
	if err := c.do(ctx, token, req, &respBody); err != nil {
		return err
	}

	// If the GraphQL response has errors, still attempt to unmarshal the data portion, as some
	// requests may expect errors but have useful responses (e.g., querying a list of repositories,
	// some of which you expect to 404).
	if len(respBody.Errors) > 0 {
		err = respBody.Errors
	}
	if result != nil && respBody.Data != nil {
		if err0 := unmarshal(respBody.Data, result); err0 != nil && err == nil {
			return err0
		}
	}
	return err
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
	if e, ok := errors.Cause(err).(*APIError); ok {
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
	if e, ok := errors.Cause(err).(*APIError); ok {
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
// requestGraphQL. See https://graphql.github.io/graphql-spec/June2018/#sec-Errors.
type graphqlErrors []struct {
	Message   string        `json:"message"`
	Type      string        `json:"type"`
	Path      []interface{} `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

const graphqlErrTypeNotFound = "NOT_FOUND"

func (e graphqlErrors) Error() string {
	return fmt.Sprintf("error in GraphQL response: %s", e[0].Message)
}

type disabledClient struct{}

func (t disabledClient) Do(r *http.Request) (*http.Response, error) {
	return nil, errors.New("http: github communication disabled")
}

// APIRoot returns the root URL of the API using the base URL of the GitHub instance.
func APIRoot(baseURL *url.URL) (apiURL *url.URL, githubDotCom bool) {
	if hostname := strings.ToLower(baseURL.Hostname()); hostname == "github.com" || hostname == "www.github.com" {
		// GitHub.com's API is hosted on api.github.com.
		return &url.URL{Scheme: "https", Host: "api.github.com", Path: "/"}, true
	}
	// GitHub Enterprise
	if baseURL.Path == "" || baseURL.Path == "/" {
		return baseURL.ResolveReference(&url.URL{Path: "/api/v3"}), false
	}
	return baseURL.ResolveReference(&url.URL{Path: "api"}), false
}

// ErrIncompleteResults is returned when the GitHub Search API returns an `incomplete_results: true` field in their response
var ErrIncompleteResults = errors.New("github repository search returned incomplete results. This is an ephemeral error from GitHub, so does not indicate a problem with your configuration. See https://developer.github.com/changes/2014-04-07-understanding-search-results-and-potential-timeouts/ for more information")

// ErrPullRequestAlreadyExists is when the requested GitHub Pull Request already exists.
var ErrPullRequestAlreadyExists = errors.New("GitHub pull request already exists")
