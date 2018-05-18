package saml

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// Middleware is middleware for SAML authentication, adding endpoints under the auth path prefix to
// enable the login flow an requiring login for all other endpoints.
//
// 🚨 SECURITY
var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getAuthHandler()(w, r, next, true)
		})
	},
	App: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			getAuthHandler()(w, r, next, false)
		})
	},
}

// getAuthHandler returns the auth HTTP handler to use, depending on whether the enhancedSAML
// experiment is enabled.
func getAuthHandler() func(http.ResponseWriter, *http.Request, http.Handler, bool) {
	if conf.EnhancedSAMLEnabled() {
		return authHandler2
	}
	return authHandler1
}
