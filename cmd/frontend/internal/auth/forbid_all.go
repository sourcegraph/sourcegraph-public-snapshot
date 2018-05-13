package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// ForbidAllAuthMiddleware forbids all requests. It is used when no auth provider is configured (as
// a safer default than "server is 100% public, no auth required").
func ForbidAllAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if conf.AuthProvider() == (schema.AuthProviders{}) {
			const msg = "Access to Sourcegraph is forbidden because no authentication provider is set in site configuration."
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
