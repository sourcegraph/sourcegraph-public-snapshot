package openidconnect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/csrf"
)

const stateCookieName = "sg-oidc-state"

type UserClaims struct {
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

// Middleware is middleware for OpenID Connect (OIDC) authentication, adding endpoints under the
// auth path prefix ("/.auth") to enable the login flow and requiring login for all other endpoints.
//
// The OIDC spec (http://openid.net/specs/openid-connect-core-1_0.html) describes an authentication protocol
// that involves 3 parties: the Relying Party (e.g., Sourcegraph), the OpenID Provider (e.g., Okta, OneLogin,
// or another SSO provider), and the End User (e.g., a user's web browser).
//
// This middleware implements two things: (1) the OIDC Authorization Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) and (2) Sourcegraph-specific session management
// (outside the scope of the OIDC spec). Upon successful completion of the OIDC login flow, the handler will create
// a new session and session cookie. The expiration of the session is the expiration of the OIDC ID Token.
//
// 🚨 SECURITY
var Middleware = &auth.Middleware{
	API: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleOpenIDConnectAuth(w, r, next, true)
		})
	},
	App: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleOpenIDConnectAuth(w, r, next, false)
		})
	},
}

// handleOpenIDConnectAuth performs OpenID Connect authentication (if configured) for HTTP requests,
// both API requests and non-API requests.
func handleOpenIDConnectAuth(w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	// Check the OpenID Connect auth provider configuration.
	pc := conf.AuthProvider().Openidconnect
	if pc != nil && pc.Issuer == "" {
		log15.Error("No issuer set for OpenID Connect auth provider (set the openidconnect auth provider's issuer property).")
		http.Error(w, "misconfigured OpenID Connect auth provider", http.StatusInternalServerError)
		return
	}

	// If actor is already authenticated (e.g., via access token), or no OpenID Connect auth
	// provider is configured, skip OpenID Connect auth.
	if actor.FromContext(r.Context()).IsAuthenticated() || pc == nil {
		next.ServeHTTP(w, r)
		return
	}

	// Delegate to the OpenID Connect login handler to handle the OIDC Authentication Code Flow
	// callback.
	if !isAPIRequest && strings.HasPrefix(r.URL.Path, auth.AuthURLPrefix+"/") {
		loginHandler(w, r, pc)
		return
	}

	// Unauthenticated API requests are rejected immediately (no redirect to auth flow because there
	// is no interactive user to redirect).
	if isAPIRequest {
		http.Error(w, "requires authentication", http.StatusUnauthorized)
		return
	}

	// Otherwise, an unauthenticated client making an app request should be redirected to the OpenID
	// Connect login flow.
	redirectURL := url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery, Fragment: r.URL.Fragment}
	query := url.Values(map[string][]string{"redirect": {redirectURL.String()}})
	http.Redirect(w, r, auth.AuthURLPrefix+"/login?"+query.Encode(), http.StatusFound)
}

// loginHandler handles the OIDC Authentication Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) on the Relying Party's end.
//
// 🚨 SECURITY
func loginHandler(w http.ResponseWriter, r *http.Request, pc *schema.OpenIDConnectAuthProvider) {
	provider, err := cache.get(pc.Issuer)
	if err != nil {
		log15.Error("Error getting OpenID Connect provider metadata.", "issuer", pc.Issuer, "error", err)
		http.Error(w, "unexpected error in OpenID Connect authentication provider", http.StatusInternalServerError)
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     pc.ClientID,
		ClientSecret: pc.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s%s/%s", globals.AppURL.String(), auth.AuthURLPrefix, "callback"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: pc.ClientID})

	switch strings.TrimPrefix(r.URL.Path, auth.AuthURLPrefix) {
	case "/login":
		// Endpoint that starts the Authentication Request Code Flow.

		// The state parameter is an opaque value used to maintain state between the original Authentication Request
		// and the callback. We do not record any state beyond a CSRF token used to defend against CSRF attacks against the callback.
		// We use the CSRF token created by gorilla/csrf that is used for other app endpoints as the OIDC state parameter.
		//
		// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the OIDC spec.
		state := (&authnState{CSRFToken: csrf.Token(r), Redirect: r.URL.Query().Get("redirect")}).Encode()
		http.SetCookie(w, &http.Cookie{Name: stateCookieName, Value: state, Expires: time.Now().Add(time.Minute * 15)})

		// Redirect to the OP's Authorization Endpoint for authentication. The nonce is an optional
		// string value used to associate a Client session with an ID Token and to mitigate replay attacks.
		// Whereas the state parameter is used in validating the Authentication Request
		// callback, the nonce is used in validating the response to the ID Token request.
		// We re-use the Authn request state as the nonce.
		//
		// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the OIDC spec.
		http.Redirect(w, r, oauth2Config.AuthCodeURL(state, oidc.Nonce(state)), http.StatusFound)

	case "/callback":
		// Endpoint for the OIDC Authorization Response. See http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
		ctx := r.Context()
		if authError := r.URL.Query().Get("error"); authError != "" {
			log15.Error("Authentication error returned by SSO provider", "error", authError, "description", r.URL.Query().Get("error_description"))
			http.Error(w, ssoErrMsg(authError, r.URL.Query().Get("error_description")), http.StatusUnauthorized)
			return
		}

		// Validate state parameter to prevent CSRF attacks
		stateParam := r.URL.Query().Get("state")
		if stateParam == "" {
			http.Error(w, ssoErrMsg("No OIDC state query parameter specified", ""), http.StatusBadRequest)
			return
		}
		stateCookie, err := r.Cookie(stateCookieName)
		if err == http.ErrNoCookie {
			log15.Error("No OIDC state cookie found - possible request forgery", "error", err)
			http.Error(w, ssoErrMsg("No OIDC state cookie found", "possible request forgery"), http.StatusBadRequest)
			return
		} else if err != nil {
			log15.Error("Could not read OIDC state cookie", "error", err)
			http.Error(w, "Could not read OIDC state cookie", http.StatusInternalServerError)
			return
		}
		if stateCookie.Value != stateParam {
			http.Error(w, ssoErrMsg("OIDC state parameter is incorrect", "possible request forgery"), http.StatusBadRequest)
			return
		}

		// Decode state param value
		var state authnState
		if err := state.Decode(stateParam); err != nil {
			log15.Error("OIDC state parameter was invalid", "error", err)
			http.Error(w, ssoErrMsg("OIDC state parameter was invalid", ""), http.StatusBadRequest)
			return
		}

		// Exchange the code for an access token. See http://openid.net/specs/openid-connect-core-1_0.html#TokenRequest.
		oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			log15.Error("Failed to obtain access token from OpenID Provider", "error", err)
			http.Error(w, ssoErrMsg("Failed to obtain access token from OpenID Provider", err), http.StatusUnauthorized)
			return
		}

		// Extract the ID Token from the Access Token. See http://openid.net/specs/openid-connect-core-1_0.html#TokenResponse.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, ssoErrMsg("Authorization response did not contain ID Token", ""), http.StatusUnauthorized)
			return
		}

		// Parse and verify ID Token payload. See http://openid.net/specs/openid-connect-core-1_0.html#TokenResponseValidation.
		var idToken *oidc.IDToken
		if mockVerifyIDToken != nil {
			idToken = mockVerifyIDToken(rawIDToken)
		} else {
			idToken, err = verifier.Verify(ctx, rawIDToken)
			if err != nil {
				log15.Error("ID Token verification failed", "error", err)
				http.Error(w, ssoErrMsg("ID Token verification failed", ""), http.StatusUnauthorized)
				return
			}
		}

		// Validate the nonce. The Verify method explicitly doesn't handle nonce validation, so we do that here.
		// We set the nonce to be the same as the state in the Authentication Request state, so we check for equality
		// here.
		if idToken.Nonce != stateParam {
			http.Error(w, ssoErrMsg("Incorrect nonce value when fetching ID Token, possible replay attack", ""), http.StatusUnauthorized)
			return
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			log15.Error("Failed to get userinfo", "error", err)
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if pc.RequireEmailDomain != "" && !strings.HasSuffix(userInfo.Email, "@"+pc.RequireEmailDomain) {
			http.Error(w, ssoErrMsg("Invalid email domain", "Required: "+pc.RequireEmailDomain), http.StatusUnauthorized)
			return
		}

		var claims UserClaims
		if err := userInfo.Claims(&claims); err != nil {
			log15.Warn("Could not parse userInfo claims", "error", err)
		}
		actr, err := getActor(ctx, idToken, userInfo, &claims)
		if err != nil {
			log15.Error("Error looking up OpenID-authenticated user.", "error", err)
			http.Error(w, "Error looking up OpenID-authenticated user. "+auth.CouldNotGetUserDescription, http.StatusInternalServerError)
			return
		}

		var exp time.Duration
		if !idToken.Expiry.IsZero() {
			exp = time.Until(idToken.Expiry)
		}
		if err := session.SetActor(w, r, actr, exp); err != nil {
			log15.Error("Could not initiate session", "error", err)
			http.Error(w, ssoErrMsg("Could not initiate session", err), http.StatusInternalServerError)
			return
		}

		data := sessionData{
			Issuer:   pc.Issuer,
			ClientID: pc.ClientID,
			Token:    oauth2Token,
		}
		if err := session.SetData(w, r, sessionKey, data); err != nil {
			log15.Error("Failed to set OpenID Connect session data.", "error", err)
			http.Error(w, "Unexpected error initiating OpenID Connect session.", http.StatusInternalServerError)
			return
		}

		// 🚨 SECURITY: Call auth.SafeRedirectURL to avoid an open-redirect vuln.
		http.Redirect(w, r, auth.SafeRedirectURL(state.Redirect), http.StatusFound)

	default:
		http.Error(w, "", http.StatusNotFound)
	}
}

// getActor returns the actor corresponding to the user indicated by the OIDC ID Token and UserInfo response.
// Because Actors must correspond to users in our DB, it creates the user in the DB if the user does not yet
// exist.
func getActor(ctx context.Context, idToken *oidc.IDToken, userInfo *oidc.UserInfo, claims *UserClaims) (*actor.Actor, error) {
	provider := idToken.Issuer
	externalID := oidcToExternalID(provider, idToken.Subject)
	login := claims.PreferredUsername
	if login == "" {
		login = userInfo.Email
	}
	email := userInfo.Email
	var displayName = claims.GivenName
	if displayName == "" {
		if claims.Name == "" {
			displayName = claims.Name
		} else {
			displayName = login
		}
	}
	login, err := auth.NormalizeUsername(login)
	if err != nil {
		return nil, err
	}

	userID, err := auth.CreateOrUpdateUser(ctx, db.NewUser{
		Username:    login,
		Email:       email,
		DisplayName: displayName,
		AvatarURL:   claims.Picture,
	}, db.ExternalAccountSpec{ServiceType: "openidconnect", ServiceID: idToken.Issuer, AccountID: externalID})
	if err != nil {
		return nil, err
	}
	return actor.FromUser(userID), nil
}

func ssoErrMsg(err string, description interface{}) string {
	return fmt.Sprintf("SSO error: %s\n%v", err, description)
}

// mockVerifyIDToken mocks the OIDC ID Token verification step. It should only be set in tests.
var mockVerifyIDToken func(rawIDToken string) *oidc.IDToken

// authnState is the state parameter passed to the Authn request and returned in the Authn response callback.
type authnState struct {
	CSRFToken string `json:"csrfToken"`
	Redirect  string `json:"redirect"`
}

// Encode returns the base64-encoded JSON representation of the authn state.
func (s *authnState) Encode() string {
	b, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the base64-encoded JSON representation of the authn state into the receiver.
func (s *authnState) Decode(encoded string) error {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, s)
}

func oidcToExternalID(issuer, subject string) string {
	return fmt.Sprintf("%s:%s", issuer, subject)
}
