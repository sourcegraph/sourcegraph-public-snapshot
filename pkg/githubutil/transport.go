package githubutil

import (
	"net/http"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
)

func NewGitHubCacheControlTransport(origCacheControl string, baseTransport http.RoundTripper) *httputil.CacheControlTransport {
	var cacheControl string
	switch strings.ToLower(origCacheControl) {
	case "max-age=0":
		fallthrough
	case "no-cache":
		// grace period of 2 seconds
		cacheControl = "max-age=2"
	}

	return httputil.NewCacheControlTransport(cacheControl, baseTransport, func(r *http.Request) bool {
		if strings.HasSuffix(r.URL.Path, "/comments") || strings.Contains(r.URL.Path, "/comments/") || strings.HasSuffix(r.URL.Path, "/pulls") || strings.Contains(r.URL.Path, "/pulls/") {
			return true
		}
		return false
	})
}
