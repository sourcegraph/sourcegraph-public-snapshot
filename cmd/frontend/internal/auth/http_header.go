package auth

import (
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
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
// TESTING: Use the testproxy test program to test HTTP auth proxy behavior. For example, run `go
// run cmd/frontend/internal/auth/testproxy.go -username=alice` then go to
// http://localhost:4080. See `-h` for flag help.
//
// 🚨 SECURITY
func newHTTPHeaderAuthHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the HTTP request contains the header from the auth proxy, get or create the user and
		// proceed with the authenticated request.
		if headerValue := r.Header.Get(ssoUserHeader); headerValue != "" {
			username, err := NormalizeUsername(headerValue)
			if err != nil {
				log15.Error("Error normalizing username from HTTP auth proxy.", "username", headerValue, "err", err)
				http.Error(w, "unable to normalize username", http.StatusInternalServerError)
				return
			}
			userID, err := createOrUpdateUser(r.Context(), db.NewUser{
				ExternalProvider: UserProviderHTTPHeader,
				// Store headerValue, not normalized username, to prevent two users with distinct
				// pre-normalization usernames from being merged into the same normalized username
				// (and therefore letting them each impersonate the other).
				ExternalID: UserProviderHTTPHeader + ":" + headerValue,
				Username:   username,
			})
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
