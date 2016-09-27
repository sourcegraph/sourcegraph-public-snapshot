package handlerutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	"github.com/gorilla/csrf"
)

type Middleware func(next http.Handler) http.Handler

func WithMiddleware(h http.Handler, mw ...Middleware) http.Handler {
	if len(mw) == 0 {
		return h
	}
	return mw[0](WithMiddleware(h, mw[1:]...))
}

// NewHandlerWithCSRFProtection creates a new handler that uses the provided
// handler. It additionally adds support for cross-site request forgery. To make
// your forms compliant you will have to include a hidden input which contains
// the CSRFToken that is made available to you in the template via tmpl.Common.
//
// Example:
// 	<input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
//
func NewHandlerWithCSRFProtection(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := csrf.Protect(
			[]byte("e953612ddddcdd5ec60d74e07d40218c"),
			csrf.CookieName("csrf_token"),
			csrf.Path("/"),
			csrf.Secure(conf.AppURL.Scheme == "https"),
		)
		p(handler).ServeHTTP(w, r)
	})
}
