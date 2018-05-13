package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const overrideHeader = "X-Override-Auth-Secret"

// OverrideAuthMiddleware is middleware that causes a new authenticated session (associated with a
// new user named "anon-user") to be started if the client provides a secret header value specified
// in site config.
//
// It is used to enable our e2e tests to authenticate to https://sourcegraph.sgdev.org without
// needing to give them G Suite access.
func OverrideAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := getOverrideAuthSecret()
		// Accept both old header (X-Oidc-Override, deprecated) and new overrideHeader for now.
		if secret != "" && (r.Header.Get("X-Oidc-Override") == secret || r.Header.Get(overrideHeader) == secret) {
			userID, err := CreateOrUpdateUser(r.Context(), db.NewUser{
				ExternalProvider: "override",
				ExternalID:       "anon-user",
				Username:         "anon-user",
				Email:            "anon-user@sourcegraph.com",
			})
			if err != nil {
				log15.Error("Error getting/creating anonymous user.", "error", err)
				http.Error(w, "error getting/creating anonymous user", http.StatusInternalServerError)
				return
			}

			a := actor.FromUser(userID)
			if err := session.StartNewSession(w, r, a, 0); err != nil {
				log15.Error("Error starting anonymous session.", "error", err)
				http.Error(w, "error starting anonymous session", http.StatusInternalServerError)
				return
			}

			r = r.WithContext(actor.WithActor(r.Context(), a))
		}

		next.ServeHTTP(w, r)
	})
}

// envOverrideAuthSecret (the env var OVERRIDE_AUTH_SECRET) is the preferred source of the secret
// for overriding auth.
var envOverrideAuthSecret = env.Get("OVERRIDE_AUTH_SECRET", "", "X-Override-Auth-Secret HTTP request header value used to create authed sessions in e2e tests")

func getOverrideAuthSecret() string {
	if envOverrideAuthSecret != "" {
		return envOverrideAuthSecret
	}
	if c := conf.Get(); c.AuthOpenIDConnect != nil {
		return c.AuthOpenIDConnect.OverrideToken
	}
	return ""
}
