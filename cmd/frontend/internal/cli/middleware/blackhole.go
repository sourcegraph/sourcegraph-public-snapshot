package middleware

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
)

// BlackHole is a middleware which returns StatusGone on removed URLs that
// external systems still regularly hit
func BlackHole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBlackhole(r) {
			next.ServeHTTP(w, r)
			return
		}

		httptrace.SetRouteName(r, "middleware.blackhole")
		w.WriteHeader(http.StatusGone)
	})
}

func isBlackhole(r *http.Request) bool {
	// We no longer support github webhooks
	if r.URL.Path == "/api/ext/github/webhook" {
		return true
	}

	// We no longer support gRPC
	if r.Header.Get("content-type") == "application/grpc" {
		return true
	}

	return false
}
