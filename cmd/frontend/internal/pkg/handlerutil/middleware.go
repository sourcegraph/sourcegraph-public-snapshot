package handlerutil

import (
	"net/http"

	"github.com/gorilla/csrf"
)

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
