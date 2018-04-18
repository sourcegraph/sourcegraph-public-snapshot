package httpapi

import (
	"net/http"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

// authorizationMiddleware authenticates the user based on the "Authorization" header.
func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		switch strings.ToLower(parts[0]) {
		case "token":
			if conf.AccessTokensEnabled() {
				userID, err := db.AccessTokens.Lookup(r.Context(), parts[1])
				if err != nil {
					log15.Error("Invalid access token.", "token", parts[1], "err", err)
					http.Error(w, "invalid access token", http.StatusUnauthorized)
					return
				}
				r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: userID}))
			}
		}

		next.ServeHTTP(w, r)
	})
}
