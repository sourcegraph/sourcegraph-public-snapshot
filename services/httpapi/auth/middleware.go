package auth

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

const oauthBasicUsername = "x-oauth-basic"

// AuthorizationMiddleware authenticates the user based on the "Authorization" header.
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept, Authorization, Cookie")

		if username, password, ok := r.BasicAuth(); ok {
			switch username {
			case "x-oauth-basic":
				// Allow token to be specified using HTTP Basic auth with username "x-oauth-basic".
				if len(password) == 255 {
					log.Println("WARNING: The server received an OAuth2 access token that is exactly 255 characters long. This may indicate that the client's version of libcurl is older than 7.33.0 and does not support longer passwords in HTTP Basic auth. Sourcegraph's access tokens may exceed 255 characters, in which case libcurl will truncate them and auth will fail. If you notice auth failing, try upgrading both the OpenSSL and GnuTLS flavours of libcurl to a version 7.33.0 or greater. If that doesn't solve the issue, please report it.")
				}
				r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), password))

			default:
				cl := handlerutil.Client(r)
				// Request access token based on username and password.
				tok, err := cl.Auth.GetAccessToken(r.Context(), &sourcegraph.AccessTokenRequest{
					AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
						ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: username, Password: password},
					},
				})
				if err != nil {
					log.Printf("PasswordMiddleware: error getting resource owner password access token for user %q: %s.", username, err)
					http.Error(w, "error getting access token for username/password", http.StatusForbidden)
					return
				}
				r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), tok.AccessToken))
			}
		}

		for _, h := range r.Header["Authorization"] {
			if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
				r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), h[7:]))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func AuthorizationHeader(ctx context.Context) string {
	accessToken := sourcegraph.AccessTokenFromContext(ctx)
	if accessToken == "" {
		return ""
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte("x-oauth-basic:"+accessToken))
}
