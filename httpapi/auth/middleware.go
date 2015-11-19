package auth

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"strings"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

const oauthBasicUsername = "x-oauth-basic"

// PasswordMiddleware configures API calls to use an access token
// based on the HTTP Basic credentials in the request (if any).
func PasswordMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	username, password, ok := r.BasicAuth()
	if ok && username != oauthBasicUsername {
		ctx := httpctx.FromRequest(r)

		// Request access token based on username and password.
		tok, err := handlerutil.APIClient(r).Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
			AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
				ResourceOwnerPassword: &sourcegraph.LoginCredentials{Login: username, Password: password},
			},
			TokenURL: oauth2client.TokenURL(),
		})
		if err != nil {
			log.Printf("PasswordMiddleware: error getting resource owner password access token for user %q: %s.", username, err)
			http.Error(w, "error getting access token for username/password", http.StatusForbidden)
			return
		}
		accessTok, err := getOAuthAccessToken(ctx, tok.AccessToken)
		if err != nil {
			log.Printf("PasswordMiddleware: error getting oauth access token for user %q: %s.", username, err)
			http.Error(w, "error getting oauth access token for username/password", http.StatusForbidden)
			return
		}
		ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessTok, TokenType: "Bearer"}))
		httpctx.SetForRequest(r, ctx)

		// Vary based on Authorization header if the request is
		// operating with any level of authorization, so that the
		// response can't be cached and mixed in with unauthorized
		// responses in an HTTP cache.
		w.Header().Add("vary", "Authorization")
	}
	next(w, r)
}

// OAuth2AccessTokenMiddleware configures API calls to use an OAuth2
// Bearer access token from the HTTP request (if any).
func OAuth2AccessTokenMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	tok, tokType, err := readBearerToken(r)
	if err != nil {
		log.Printf("OAuth2TokenMiddleware: error in readBearerToken: %s.", err)
		http.Error(w, "error reading access token from HTTP request", http.StatusForbidden)
		return
	}

	if len(tok) == 255 {
		log.Println("WARNING: The server received an OAuth2 access token that is exactly 255 characters long. This may indicate that the client's version of git and/or curl is old and does not support longer passwords in HTTP Basic auth. Sourcegraph's access tokens may exceed 255 characters, in which case git/curl will truncate them and auth will fail. If you notice auth failing, try upgrading the git/curl versions. If that doesn't solve the issue, please report it.")
	}

	if tok != "" {
		ctx := httpctx.FromRequest(r)
		ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok, TokenType: tokType}))
		httpctx.SetForRequest(r, ctx)

		// Vary based on Authorization header if the request is
		// operating with any level of authorization, so that the
		// response can't be cached and mixed in with unauthorized
		// responses in an HTTP cache.
		w.Header().Add("vary", "Authorization")
	}
	next(w, r)
}

func readBearerToken(r *http.Request) (token, tokenType string, err error) {
	// Allow token to be specified using HTTP Basic auth with username
	// "x-oauth-basic".
	username, token, ok := r.BasicAuth()
	if ok && username == oauthBasicUsername {
		return token, "Bearer", nil
	}

	for _, v := range r.Header[http.CanonicalHeaderKey("authorization")] {
		parts := strings.SplitN(v, " ", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(parts[0], "bearer") {
			return parts[1], "Bearer", nil
		}
	}

	return "", "", nil
}

// getOAuthAccessToken returns the OAuth access token for authenticating
// the user on the local server. It fetches an auth code from the root
// and exchanges it for an access token.
func getOAuthAccessToken(ctx context.Context, accessTok string) (string, error) {
	if fed.Config.IsRoot || authutil.ActiveFlags.IsLocal() || authutil.ActiveFlags.IsLDAP() {
		return accessTok, nil
	}
	rootCtx := fed.NewRemoteContext(ctx, fed.Config.RootURL())
	rootCl := sourcegraph.NewClientFromContext(rootCtx)
	// Use the root access token to issue an auth code that
	// this server can then exchange for an access token.
	rootAuthedCtx := sourcegraph.WithCredentials(ctx,
		oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: accessTok}),
	)
	authInfo, err := rootCl.Auth.Identify(rootAuthedCtx, &pbtypes.Void{})
	if err != nil {
		return "", fmt.Errorf("could not identify root user: %s", err)
	}

	code, err := rootCl.Auth.GetAuthorizationCode(rootAuthedCtx, &sourcegraph.AuthorizationCodeRequest{
		ResponseType: "code",
		ClientID:     idkey.FromContext(ctx).ID,
		UID:          authInfo.UID,
	})
	if err != nil {
		return "", fmt.Errorf("could not get auth code from root: %s", err)
	}

	// Exchange the auth code (from the root) for an access token.
	tok, err := sourcegraph.NewClientFromContext(ctx).Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_AuthorizationCode{
			AuthorizationCode: code,
		},
		TokenURL: oauth2client.TokenURL(),
	})
	if err != nil {
		return "", fmt.Errorf("could not exchange auth code for access token: %s", err)
	}

	return tok.AccessToken, nil
}
