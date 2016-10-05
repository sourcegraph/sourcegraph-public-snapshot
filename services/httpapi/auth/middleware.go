package auth

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

const oauthBasicUsername = "x-oauth-basic"

var (
	legacyAuthCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "auth",
		Name:      "legacy_auth",
		Help:      "The number of legacy authentications.",
	})
)

func init() {
	prometheus.MustRegister(legacyAuthCounter)
}

// AuthorizationMiddleware authenticates the user based on the "Authorization" header.
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept, Authorization, Cookie")

		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 {
			r = r.WithContext(github.NewContextWithAuthedClient(r.Context()))
			next.ServeHTTP(w, r)
			return
		}

		switch strings.ToLower(parts[0]) {
		case "basic":
			username, password, _ := r.BasicAuth()
			switch username {
			case "x-oauth-basic": // Legacy: Allow token to be specified using HTTP Basic auth with username "x-oauth-basic".
				legacyAuthCounter.Inc()
				accessToken := password
				if len(accessToken) == 255 {
					log.Println("WARNING: The server received an OAuth2 access token that is exactly 255 characters long. This may indicate that the client's version of libcurl is older than 7.33.0 and does not support longer passwords in HTTP Basic auth. Sourcegraph's access tokens may exceed 255 characters, in which case libcurl will truncate them and auth will fail. If you notice auth failing, try upgrading both the OpenSSL and GnuTLS flavours of libcurl to a version 7.33.0 or greater. If that doesn't solve the issue, please report it.")
				}
				if _, err := auth.ParseAndVerify(accessToken); err == nil { // Legacy support for Chrome extension: This might be a session cookie.
					r = r.WithContext(auth.AuthenticateByAccessToken(r.Context(), accessToken))
				} else {
					r = r.WithContext(auth.AuthenticateBySession(r.Context(), password))
				}
			}

		case "token", "bearer":
			r = r.WithContext(auth.AuthenticateByAccessToken(r.Context(), parts[1]))

		case "session":
			r = r.WithContext(auth.AuthenticateBySession(r.Context(), parts[1]))

		}

		r = r.WithContext(github.NewContextWithAuthedClient(r.Context()))
		next.ServeHTTP(w, r)
	})
}

// AuthorizationHeaderWithAccessToken returns a value for the "Authorization" header that can be
// used to authenticate the current user. This should only be used internally since the access
// token can not be revoked.
func AuthorizationHeaderWithAccessToken(ctx context.Context) string {
	a := auth.ActorFromContext(ctx)
	if !a.IsAuthenticated() {
		return ""
	}
	return "token " + auth.NewAccessToken(a, time.Hour)
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
