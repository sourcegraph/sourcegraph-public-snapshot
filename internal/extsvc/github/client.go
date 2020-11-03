package github

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

// RateLimitMonitor exposes the rate limit monitor.
func (c *V3Client) RateLimitMonitor() *ratelimit.Monitor {
	return c.rateLimitMonitor
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
