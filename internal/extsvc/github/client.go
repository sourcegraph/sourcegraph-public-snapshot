package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

// NewClient creates a new GitHub API client with an optional default
// authenticator.
//
// apiURL must point to the base URL of the GitHub API. See the docstring for
// Client.apiURL.
func NewClient(apiURL *url.URL, a auth.Authenticator, cli httpcli.Doer) *V3Client {
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

	rl := ratelimit.DefaultRegistry.Get(apiURL.String())

	return &V3Client{
		apiURL:           apiURL,
		githubDotCom:     urlIsGitHubDotCom(apiURL),
		auth:             a,
		httpClient:       cli,
		rateLimitMonitor: &ratelimit.Monitor{HeaderPrefix: "X-"},
		repoCache:        newRepoCache(apiURL, a),
		rateLimit:        rl,
	}
}

// WithAuthenticator returns a new Client that uses the same configuration as
// the current Client, except authenticated as the GitHub user with the given
// authenticator instance (most likely a token).
func (c *V3Client) WithAuthenticator(a auth.Authenticator) *V3Client {
	return NewClient(c.apiURL, a, c.httpClient)
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
		return errors.Wrap(err, "rate limit")
	}

	return doRequest(ctx, c.apiURL, c.auth, c.rateLimitMonitor, c.httpClient, req, result)
}

// RateLimitMonitor exposes the rate limit monitor.
func (c *V3Client) RateLimitMonitor() *ratelimit.Monitor {
	return c.rateLimitMonitor
}

// HTTPErrorCode returns err's HTTP status code, if it is an HTTP error from
// this package. Otherwise it returns 0.
func HTTPErrorCode(err error) int {
	if e, ok := errors.Cause(err).(*APIError); ok {
		return e.Code
	}
	return 0
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

// ErrPullRequestAlreadyExists is thrown when the requested GitHub Pull Request already exists.
var ErrPullRequestAlreadyExists = errors.New("GitHub pull request already exists")

// ErrPullRequestNotFound is when the requested GitHub Pull Request doesn't exist.
type ErrPullRequestNotFound int

func (e ErrPullRequestNotFound) Error() string {
	return fmt.Sprintf("GitHub pull requests not found: %d", e)
}
