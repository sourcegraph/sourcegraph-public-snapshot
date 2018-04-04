package auth

import (
	"net/http"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

// AuthorizationMiddleware authenticates the user based on the "Authorization" header.
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept, Authorization, Cookie")

		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		switch strings.ToLower(parts[0]) {
		case "session":
			r = r.WithContext(session.AuthenticateBySession(r.Context(), parts[1]))
		case "token":
			userID, err := db.AccessTokens.Lookup(r.Context(), parts[1])
			if err != nil {
				log15.Error("Invalid access token.", "token", parts[1], "err", err)
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}
			r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: userID}))
		}

		next.ServeHTTP(w, r)
	})
}

// AuthorizationHeaderWithSessionCookie returns a value for the "Authorization" header that can be
// used to authenticate the current user. This header can be sent to the client, but is a bit more
// expensive to verify.
func AuthorizationHeaderWithSessionCookie(sessionCookie string) string {
	if sessionCookie == "" {
		return ""
	}
	return "session " + sessionCookie
}
