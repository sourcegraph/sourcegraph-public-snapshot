package handlerutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/gorilla/csrf"
)

var passwordEnv = env.Get("PASSWORD", "", "password for basic authentication")

// NewBasicAuthHandler creates a new handler that wraps an existing handler
// with HTTP basic authentication.
func NewBasicAuthHandler(handler http.Handler) http.Handler {
	return NewBasicAuthHandlerWithPassword(handler, passwordEnv)
}

func NewBasicAuthHandlerWithPassword(handler http.Handler, expectedPassword string) http.Handler {
	if expectedPassword == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, password, _ := r.BasicAuth(); password != expectedPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="All"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// NewHandlerWithCSRFProtection creates a new handler that uses the provided
// handler. It additionally adds support for cross-site request forgery. To make
// your forms compliant you will have to submit the token via the X-Csrf-Token
// header, which is made available in the client-side context.
func NewHandlerWithCSRFProtection(handler http.Handler, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := csrf.Protect(
			[]byte("e953612ddddcdd5ec60d74e07d40218c"),
			csrf.CookieName("csrf_token"),
			csrf.Path("/"),
			csrf.Secure(secure),
		)
		p(handler).ServeHTTP(w, r)
	})
}
