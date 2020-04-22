// Package hooks allow hooking into the frontend.
package hooks

import "net/http"

// PostAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is authenticated when PostAuthMiddleware is called.
var PostAuthMiddleware func(http.Handler) http.Handler
