// Package redirects, when imported, adds a middleware to the app that
// redirects from a list of hardcoded old URLs to new URLs.
package redirects

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
)

func init() {
	internal.Middleware = append(internal.Middleware, redirectsMiddleware)
}

// redirects is a mapping from old URL path to new destination
// URL. Note that map keys are URL paths, not full URLs, so (e.g.) a
// map key of "/path" will match request URIs of "/path" and
// "/path?a=b".
var redirects = map[string]string{}

// redirectsMiddleware sends an HTTP 301 Moved Permanently response
// with the new destination URL if the request URL's path is in the
// redirects map.
func redirectsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dest, present := redirects[r.URL.Path]; present {
			http.Redirect(w, r, dest, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
