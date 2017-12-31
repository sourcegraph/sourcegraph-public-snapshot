package auth

import (
	"context"
	"net/http"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

var ssoUserHeader = conf.AuthHTTPHeader()

// AuthorizationMiddleware authenticates the user based on the "Authorization" header.
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept, Authorization, Cookie")

		if ssoUserHeader != "" {
			if username := r.Header.Get(ssoUserHeader); username != "" {
				userID, err := getUserFromSSOHeaderUsername(r.Context(), username)
				if err != nil {
					log15.Error("unable to get/create user from SSO header", "header", ssoUserHeader, "username", username, "err", err)
					http.Error(w, "unable to get/create user", http.StatusInternalServerError)
					return
				}

				r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: userID}))
				next.ServeHTTP(w, r)
				return
			}
		}

		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		switch strings.ToLower(parts[0]) {
		case "session":
			r = r.WithContext(session.AuthenticateBySession(r.Context(), parts[1]))
		}

		next.ServeHTTP(w, r)
	})
}

func getUserFromSSOHeaderUsername(ctx context.Context, username string) (userID int32, err error) {
	user, err := db.Users.GetByUsername(ctx, username)
	if err == nil {
		// User exists.
		return user.ID, nil
	} else if _, ok := err.(db.ErrUserNotFound); !ok {
		return 0, err
	}

	// User does not exist, so we need to create it.
	user, err = db.Users.Create(ctx, sourcegraph.UserProviderHTTPHeader+":"+username, "", username, "", sourcegraph.UserProviderHTTPHeader, "", "")
	// Handle the race condition where the new user performs two requests
	// and both try to create the user.
	if err == db.ErrUsernameExists || err == db.ErrAuthIDExists {
		user, err = db.Users.GetByUsername(ctx, username)
	}
	if err != nil {
		return 0, err
	}
	return user.ID, nil
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
