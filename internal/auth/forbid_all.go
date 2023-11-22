package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// ForbidAllRequestsMiddleware forbids all requests. It is used when no auth provider is configured (as
// a safer default than "server is 100% public, no auth required").
func ForbidAllRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(conf.Get().AuthProviders) == 0 {
			const msg = "Access to Sourcegraph is forbidden because no authentication provider is set in site configuration."
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
