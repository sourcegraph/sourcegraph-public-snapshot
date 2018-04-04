package auth

import (
	"context"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

const (
	// UserProviderHTTPHeader is the http-header auth provider.
	UserProviderHTTPHeader = "http-header"
)

var ssoUserHeader = conf.AuthHTTPHeader()

// newHTTPHeaderAuthHandler wraps the handler and checks for an HTTP header from an auth proxy that
// specifies the client's authenticated username. It's for use with auth proxies like
// https://github.com/bitly/oauth2_proxy and is configured with the auth.provider=="http-header"
// site config setting.
//
// 🚨 SECURITY
func newHTTPHeaderAuthHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the HTTP request contains the header from the auth proxy, get or create the user and
		// proceed with the authenticated request.
		if headerValue := r.Header.Get(ssoUserHeader); headerValue != "" {
			userID, err := getUserFromSSOHeaderUsername(r.Context(), headerValue)
			if err != nil {
				log15.Error("unable to get/create user from SSO header", "header", ssoUserHeader, "headerValue", headerValue, "err", err)
				http.Error(w, "unable to get/create user", http.StatusInternalServerError)
				return
			}

			r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: userID}))
			handler.ServeHTTP(w, r)
			return
		}

		// The auth proxy is expected to proxy *all* requests, so don't let any non-proxied requests
		// proceed. Note that if an attacker can send HTTP requests directly to this server (not via
		// the proxy), then the attacker can impersonate any user by just sending a request with a
		// certain easy-to-construct header, so this doesn't actually provide any security.
		http.Error(w, "must access via HTTP authentication proxy", http.StatusUnauthorized)
	})

}

func getUserFromSSOHeaderUsername(ctx context.Context, username string) (userID int32, err error) {
	username, err = NormalizeUsername(username)
	if err != nil {
		return 0, err
	}

	user, err := db.Users.GetByUsername(ctx, username)
	if err == nil {
		// User exists.
		return user.ID, nil
	} else if !errcode.IsNotFound(err) {
		return 0, err
	}

	// User does not exist, so we need to create it.
	user, err = db.Users.Create(ctx, db.NewUser{
		ExternalID:       UserProviderHTTPHeader + ":" + username,
		Username:         username,
		ExternalProvider: UserProviderHTTPHeader,
	})
	// Handle the race condition where the new user performs two requests
	// and both try to create the user.
	if err != nil {
		var err2 error
		user, err2 = db.Users.GetByUsername(ctx, username)
		if err2 != nil {
			return 0, err // return Create error
		}
	}
	return user.ID, nil
}
