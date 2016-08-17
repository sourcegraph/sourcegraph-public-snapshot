package middleware

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

// BlackHole is a middleware which returns StatusGone on removed URLs that
// external systems still regularly hit
func BlackHole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ext/github/webhook" {
			next.ServeHTTP(w, r)
			return
		}

		r = httpctx.WithRouteName(r, "blackhole")
		w.WriteHeader(http.StatusGone)
	})
}
