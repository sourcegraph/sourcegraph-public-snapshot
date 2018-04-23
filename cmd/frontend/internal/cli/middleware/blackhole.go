package middleware

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// BlackHole is a middleware which returns StatusGone on removed URLs that
// external systems still regularly hit.
//
// 🚨 SECURITY: This handler is served to all clients, even on private servers to clients who have
// not authenticated. It must not reveal any sensitive information.
func BlackHole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBlackhole(r) {
			next.ServeHTTP(w, r)
			return
		}

		trace.SetRouteName(r, "middleware.blackhole")
		w.WriteHeader(http.StatusGone)
	})
}

func isBlackhole(r *http.Request) bool {
	// We no longer support github webhooks
	if r.URL.Path == "/api/ext/github/webhook" || r.URL.Path == "/.api/github-webhooks" {
		return true
	}

	// We no longer support gRPC
	if r.Header.Get("content-type") == "application/grpc" {
		return true
	}

	return false
}
