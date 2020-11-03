package github

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

var (
	gitHubDisable, _ = strconv.ParseBool(env.Get("SRC_GITHUB_DISABLE", "false", "disables communication with GitHub instances. Used to test GitHub service degradation"))
	githubProxyURL   = func() *url.URL {
		url, err := url.Parse(env.Get("GITHUB_BASE_URL", "http://github-proxy", "base URL for GitHub.com API (used for github-proxy)"))
		if err != nil {
			log.Fatal("Error parsing GITHUB_BASE_URL:", err)
		}
		return url
	}()

	requestCounter = metrics.NewRequestMeter("github", "Total number of requests sent to the GitHub API.")
)

func canonicalizedURL(apiURL *url.URL) *url.URL {
	if urlIsGitHubDotCom(apiURL) {
		// For GitHub.com API requests, use github-proxy (which adds our OAuth2 client ID/secret to get a much higher
		// rate limit).
		return githubProxyURL
	}
	return apiURL
}

func urlIsGitHubDotCom(apiURL *url.URL) bool {
	hostname := strings.ToLower(apiURL.Hostname())
	return hostname == "api.github.com" || hostname == "github.com" || hostname == "www.github.com" || apiURL.String() == githubProxyURL.String()
}

// ErrNotFound is when the requested GitHub repository is not found.
var ErrNotFound = errors.New("GitHub repository not found")

// IsNotFound reports whether err is a GitHub API error of type NOT_FOUND, the equivalent cached
// response error, or HTTP 404.
func IsNotFound(err error) bool {
	if err == ErrNotFound || errors.Cause(err) == ErrNotFound {
		return true
	}
	if _, ok := err.(ErrPullRequestNotFound); ok {
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
