package authutil

import (
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// ConnectOrSignOutMiddleware determines whether or not an incoming
// auth request is a sign-in or account connection attempt.
// If it is a sign-in attempt, it signs the user out by clearing their session.
func ConnectOrSignOutMiddleware(db database.DB) *auth.Middleware {
	return &auth.Middleware{
		API: func(next http.Handler) http.Handler {
			return next
		},

		App: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/login") {
					if err := connectOrSignOut(w, r); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				next.ServeHTTP(w, r)
			})
		},
	}
}

// connectOrSignOut checks whether or not the external account sign-in
// is a sign-in or an account connection attempt. If it a sign-in,
// it signs the user out by clearing their session.
func connectOrSignOut(w http.ResponseWriter, r *http.Request) error {
	isConnect := r.URL.Query().Get("connect") == "true"

	if !isConnect && actor.FromContext(r.Context()).IsAuthenticated() {
		err := session.SetActor(w, r, nil, 0, time.Time{})
		if err != nil {
			return err
		}
	}

	return nil
}
