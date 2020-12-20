package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"golang.org/x/time/rate"
)

// V3Client is a caching GitHub API client for GitHub's REST API v3.
//
// All instances use a map of rcache.Cache instances for caching (see the `repoCache` field). These
// separate instances have consistent naming prefixes so that different instances will share the
// same Redis cache entries (provided they were computed with the same API URL and access
// token). The cache keys are agnostic of the http.RoundTripper transport.
type V3Client struct {
	// apiURL is the base URL of a GitHub API. It must point to the base URL of the GitHub API. This
	// is https://api.github.com for GitHub.com and http[s]://[github-enterprise-hostname]/api for
	// GitHub Enterprise.
	apiURL *url.URL

	// githubDotCom is true if this client connects to github.com.
	githubDotCom bool

	// auth is used to authenticate requests. May be empty, in which case the
	// default behavior is to make unauthenticated requests.
	// 🚨 SECURITY: Should not be changed after client creation to prevent
	// unauthorized access to the repository cache. Use `WithAuthenticator` to
	// create a new client with a different authenticator instead.
	auth auth.Authenticator

	// httpClient is the HTTP client used to make requests to the GitHub API.
	httpClient httpcli.Doer

	// repoCache is the repository cache associated with the token.
	repoCache *rcache.Cache

	// rateLimitMonitor is the API rate limit monitor.
	rateLimitMonitor *ratelimit.Monitor

	// rateLimit is our self imposed rate limiter
	rateLimit *rate.Limiter

	// resource specifies which API this client is intended for.
	// One of 'rest' or 'search'.
	resource string
}

// NewV3Client creates a new GitHub API client with an optional default
// authenticator.
//
// apiURL must point to the base URL of the GitHub API. See the docstring for
// V3Client.apiURL.
func NewV3Client(apiURL *url.URL, a auth.Authenticator, cli httpcli.Doer) *V3Client {
	return newV3Client(apiURL, a, "rest", cli)
}

// NewV3SearchClient creates a new GitHub API client intended for use with the
// search API with an optional default authenticator.
//
// apiURL must point to the base URL of the GitHub API. See the docstring for
// V3Client.apiURL.
func NewV3SearchClient(apiURL *url.URL, a auth.Authenticator, cli httpcli.Doer) *V3Client {
	return newV3Client(apiURL, a, "search", cli)
}

func newV3Client(apiURL *url.URL, a auth.Authenticator, resource string, cli httpcli.Doer) *V3Client {
	apiURL = canonicalizedURL(apiURL)
	if gitHubDisable {
		cli = disabledClient{}
	}
	if cli == nil {
		cli = httpcli.ExternalDoer()
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

	var tokenHash string
	if a != nil {
		tokenHash = a.Hash()
	}

	rl := ratelimit.DefaultRegistry.Get(apiURL.String())
	rlm := ratelimit.DefaultMonitorRegistry.GetOrSet(apiURL.String(), tokenHash, resource, &ratelimit.Monitor{HeaderPrefix: "X-"})

	return &V3Client{
		apiURL:           apiURL,
		githubDotCom:     urlIsGitHubDotCom(apiURL),
		auth:             a,
		httpClient:       cli,
		rateLimit:        rl,
		rateLimitMonitor: rlm,
		repoCache:        newRepoCache(apiURL, a),
		resource:         resource,
	}
}

// WithAuthenticator returns a new V3Client that uses the same configuration as
// the current V3Client, except authenticated as the GitHub user with the given
// authenticator instance (most likely a token).
func (c *V3Client) WithAuthenticator(a auth.Authenticator) *V3Client {
	return newV3Client(c.apiURL, a, c.resource, c.httpClient)
}

// RateLimitMonitor exposes the rate limit monitor.
func (c *V3Client) RateLimitMonitor() *ratelimit.Monitor {
	return c.rateLimitMonitor
}

func (c *V3Client) requestGet(ctx context.Context, requestURI string, result interface{}) error {
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

	err = c.rateLimit.Wait(ctx)
	if err != nil {
		return errInternalRateLimitExceeded
	}

	return doRequest(ctx, c.apiURL, c.auth, c.rateLimitMonitor, c.httpClient, req, result)
}

// newRepoCache creates a new cache for GitHub repository metadata. The backing
// store is Redis. A checksum of the authenticator and API URL are used as a
// Redis key prefix to prevent collisions with caches for different
// authentication and API URLs.
func newRepoCache(apiURL *url.URL, a auth.Authenticator) *rcache.Cache {
	var cacheTTL time.Duration
	if urlIsGitHubDotCom(apiURL) {
		cacheTTL = 10 * time.Minute
	} else {
		// GitHub Enterprise
		cacheTTL = 30 * time.Second
	}

	key := ""
	if a != nil {
		key = a.Hash()
	}
	return rcache.NewWithTTL("gh_repo:"+key, int(cacheTTL/time.Second))
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

func (e *APIError) Unauthorized() bool {
	return e.Code == http.StatusUnauthorized
}

// HTTPErrorCode returns err's HTTP status code, if it is an HTTP error from
// this package. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	if e, ok := errors.Cause(err).(*APIError); ok {
		return e.Code
	}
	return 0
}
