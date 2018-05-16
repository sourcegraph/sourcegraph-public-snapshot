package httpheader

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const (
	// UserProviderHTTPHeader is the http-header auth provider.
	UserProviderHTTPHeader = "http-header"
)

// Middleware is middleware that checks for an HTTP header from an auth proxy that specifies the
// client's authenticated username. It's for use with auth proxies like
// https://github.com/bitly/oauth2_proxy and is configured with the auth.provider=="http-header"
// site config setting.
//
// TESTING: Use the testproxy test program to test HTTP auth proxy behavior. For example, run `go
// run cmd/frontend/internal/auth/httpheader/testproxy.go -username=alice` then go to
// http://localhost:4080. See `-h` for flag help.
//
// 🚨 SECURITY
func Middleware(next http.Handler) http.Handler {
	const misconfiguredMessage = "Misconfigured http-header auth provider."
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authProvider, multiple := getProviderConfig()
		if multiple {
			log15.Error("At most 1 HTTP header auth provider may be set in site config.")
			http.Error(w, misconfiguredMessage, http.StatusInternalServerError)
			return
		}
		if authProvider == nil {
			next.ServeHTTP(w, r)
			return
		}
		if authProvider.UsernameHeader == "" {
			log15.Error("No HTTP header set for username (set the http-header auth provider's usernameHeader property).")
			http.Error(w, "misconfigured http-header auth provider", http.StatusInternalServerError)
			return
		}

		headerValue := r.Header.Get(authProvider.UsernameHeader)
		if headerValue == "" {
			// The auth proxy is expected to proxy *all* requests, so don't let any non-proxied requests
			// proceed. Note that if an attacker can send HTTP requests directly to this server (not via
			// the proxy), then the attacker can impersonate any user by just sending a request with a
			// certain easy-to-construct header, so this doesn't actually provide any security.
			http.Error(w, "must access via HTTP authentication proxy", http.StatusUnauthorized)
			return
		}

		// Respect already authenticated actor (e.g., via access token).
		if actor.FromContext(r.Context()).IsAuthenticated() {
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, get or create the user and proceed with the authenticated request.
		username, err := auth.NormalizeUsername(headerValue)
		if err != nil {
			log15.Error("Error normalizing username from HTTP auth proxy.", "username", headerValue, "err", err)
			http.Error(w, "unable to normalize username", http.StatusInternalServerError)
			return
		}
		userID, err := auth.CreateOrUpdateUser(r.Context(), db.NewUser{Username: username}, db.ExternalAccountSpec{
			ServiceType: UserProviderHTTPHeader,

			// Store headerValue, not normalized username, to prevent two users with distinct
			// pre-normalization usernames from being merged into the same normalized username
			// (and therefore letting them each impersonate the other).
			AccountID: UserProviderHTTPHeader + ":" + headerValue,
		})
		if err != nil {
			log15.Error("unable to get/create user from SSO header", "header", authProvider.UsernameHeader, "headerValue", headerValue, "err", err)
			http.Error(w, "unable to get/create user", http.StatusInternalServerError)
			return
		}

		r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{UID: userID}))
		next.ServeHTTP(w, r)
	})
}
