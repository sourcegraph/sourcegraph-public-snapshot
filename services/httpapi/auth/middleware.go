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

// PasswordMiddleware configures API calls to use an access token
// based on the HTTP Basic credentials in the request (if any).
func PasswordMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok && username != oauthBasicUsername {
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

			// Vary based on Authorization header if the request is
			// operating with any level of authorization, so that the
			// response can't be cached and mixed in with unauthorized
			// responses in an HTTP cache.
			w.Header().Add("vary", "Authorization")
		}
		next.ServeHTTP(w, r)
	})
}

// OAuth2AccessTokenMiddleware configures API calls to use an OAuth2
// Bearer access token from the HTTP request (if any).
func OAuth2AccessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := readBearerToken(r)
		if err != nil {
			log.Printf("OAuth2TokenMiddleware: error in readBearerToken: %s.", err)
			http.Error(w, "error reading access token from HTTP request", http.StatusForbidden)
			return
		}

		if len(tok) == 255 {
			log.Println("WARNING: The server received an OAuth2 access token that is exactly 255 characters long. This may indicate that the client's version of libcurl is older than 7.33.0 and does not support longer passwords in HTTP Basic auth. Sourcegraph's access tokens may exceed 255 characters, in which case libcurl will truncate them and auth will fail. If you notice auth failing, try upgrading both the OpenSSL and GnuTLS flavours of libcurl to a version 7.33.0 or greater. If that doesn't solve the issue, please report it.")
		}

		if tok != "" {
			r = r.WithContext(sourcegraph.WithAccessToken(r.Context(), tok))

			// Vary based on Authorization header if the request is
			// operating with any level of authorization, so that the
			// response can't be cached and mixed in with unauthorized
			// responses in an HTTP cache.
			w.Header().Add("vary", "Authorization")
		}
		next.ServeHTTP(w, r)
	})
}

func readBearerToken(r *http.Request) (token string, err error) {
	// Allow token to be specified using HTTP Basic auth with username
	// "x-oauth-basic".
	username, token, ok := r.BasicAuth()
	if ok && username == oauthBasicUsername {
		return token, nil
	}

	for _, v := range r.Header[http.CanonicalHeaderKey("authorization")] {
		parts := strings.SplitN(v, " ", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(parts[0], "bearer") {
			return parts[1], nil
		}
	}

	return "", nil
}

func AuthorizationHeader(ctx context.Context) string {
	accessToken := sourcegraph.AccessTokenFromContext(ctx)
	if accessToken == "" {
		return ""
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte("x-oauth-basic:"+accessToken))
}
