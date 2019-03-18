package hooks

import "net/http"

// PreAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is not yet authenticated when PreAuthMiddleware is called.
var PreAuthMiddleware func(http.Handler) http.Handler
