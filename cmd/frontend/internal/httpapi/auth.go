package httpapi

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// AccessTokenAuthMiddleware authenticates the user based on the "Authorization" header's access
// token (if any).
func AccessTokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		if headerValue := r.Header.Get("Authorization"); headerValue != "" {
			if !conf.AccessTokensEnabled() {
				http.Error(w, "Access token authorization is disabled.", http.StatusUnauthorized)
				return
			}

			token, err := authz.ParseAuthorizationHeader(headerValue)
			if err != nil {
				log15.Error("Invalid Authorization header.", "err", err)
				http.Error(w, "Invalid Authorization header.", http.StatusUnauthorized)
				return
			}

			// Validate access token.
			subjectUserID, err := db.AccessTokens.Lookup(r.Context(), token, authz.ScopeUserAll)
			if err != nil {
				log15.Error("Invalid access token.", "token", token, "err", err)
				http.Error(w, "Invalid access token.", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: subjectUserID}))
		}

		next.ServeHTTP(w, r)
	})
}
