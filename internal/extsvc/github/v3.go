package github

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

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
