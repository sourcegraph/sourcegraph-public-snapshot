// Package redirects, when imported, adds a middleware to the app that
// redirects from a list of hardcoded old URLs to new URLs.
package redirects

import (
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
)

func init() {
	internal.Middleware = append(internal.Middleware, redirectsMiddleware)
}

// redirects is a mapping from old URL path to new destination
// URL. Note that map keys are URL paths, not full URLs, so (e.g.) a
// map key of "/path" will match request URIs of "/path" and
// "/path?a=b". Also, if the request's URL path contains a trailing
// slash, it is removed before being looked up in this map, so map
// keys should NOT have trailing slashes.
var redirects = map[string]string{
	"/blog/ipfs-the-permanent-web-by-juan-benet-talks-at":     "https://text.sourcegraph.com/ae9901b338b1",
	"/blog/why-vacation-at-tech-companies-should-be":          "https://text.sourcegraph.com/d1b549681291",
	"/blog/5-easy-ways-to-start-contributing-to-docker-using": "https://text.sourcegraph.com/211be3a30a40",
	"/blog/building-a-product-one-interview-at-a-time":        "https://text.sourcegraph.com/fb820722de13",
	"/blog/the-top-three-challenges-and-solutions-of":         "https://text.sourcegraph.com/4c3b262cae27",
	"/blog/announcing-the-sourcegraph-developer-release-the":  "https://text.sourcegraph.com/f75c298778b5",
	"/blog/getting-started-with-sourcegraph":                  "https://text.sourcegraph.com/c27af53fb24b",
	"/blog/how-to-make-your-open-source-project-thrive-with":  "https://text.sourcegraph.com/6463b935c540",
	"/blog/the-pain-of-code-review-how-different-teams":       "https://text.sourcegraph.com/f9abc79d7f3a",
	"/blog/a-url-for-every-function-in-the-world":             "https://text.sourcegraph.com/83cf36dfcddb",
	"/blog/announcing-the-sourcegraph-chrome-extension-for":   "https://text.sourcegraph.com/9e279d2b98e9",
	"/blog/announcing-appdash-an-open-source-perf-tracing":    "https://text.sourcegraph.com/4e1fc41c2031",
	"/blog/google-io-2014-building-sourcegraph-a":             "https://text.sourcegraph.com/1f911b78a82e",
}

// redirectsMiddleware sends an HTTP 301 Moved Permanently response
// with the new destination URL if the request URL's path (minus
// trailing slash) is in the redirects map.
func redirectsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" {
			path = strings.TrimSuffix(path, "/")
		}
		if dest, present := redirects[path]; present {
			http.Redirect(w, r, dest, http.StatusMovedPermanently)
			return
		}

		// Redirect all other /blog/* URLs to the main blog page.
		if r.URL.Path == "/blog" || strings.HasPrefix(r.URL.Path, "/blog/") {
			http.Redirect(w, r, "https://text.sourcegraph.com", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
