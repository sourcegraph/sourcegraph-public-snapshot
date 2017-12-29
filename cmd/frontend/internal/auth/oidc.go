package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

const oidcStateCookieName = "sg-oidc-state"

var (
	oidcProvider = conf.AuthOpenIDConnect()
)

type UserClaims struct {
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
}

// newOIDCAuthHandler wraps the passed in handler with OpenID Connect (OIDC) authentication, adding endpoints
// under the auth path prefix ("/.auth") to enable the login flow and requiring login for all other endpoints.
//
// The OIDC spec (http://openid.net/specs/openid-connect-core-1_0.html) describes an authentication protocol
// that involves 3 parties: the Relying Party (e.g., Sourcegraph Server), the OpenID Provider (e.g., Okta, OneLogin,
// or another SSO provider), and the End User (e.g., a user's web browser).
//
// The handler this method returns implements two things: (1) the OIDC Authorizaton Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) and (2) Sourcegraph-specific session management
// (outside the scope of the OIDC spec). Upon successful completion of the OIDC login flow, the handler will create
// a new session and session cookie. The expiration of the session is the expiration of the OIDC ID Token.
//
// 🚨 SECURITY
func newOIDCAuthHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	// Return an error if the OIDC parameters are unset or missing
	if oidcProvider == nil {
		return nil, errors.New("No OpenID Connect Provider specified")
	}
	if oidcProvider.ClientID == "" || oidcProvider.ClientSecret == "" {
		return nil, fmt.Errorf("OIDC Client ID or Client Secret was empty")
	}

	// Create handler for OIDC Authentication Code Flow endpoints
	oidcHandler, err := newOIDCLoginHandler(createCtx, handler, appURL)
	if err != nil {
		return nil, err
	}

	return session.CookieOrSessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if oidcProvider.OverrideToken != "" && r.Header.Get("X-Oidc-Override") == oidcProvider.OverrideToken {
			if err := startAnonUserSession(createCtx, w, r); err != nil {
				http.Error(w, "Error initializing anonymous user", http.StatusInternalServerError)
				return
			}
			handler.ServeHTTP(w, r)
			return
		}

		// If the path is under the authentication path, serve the OIDC Authentication Code Flow handler
		if strings.HasPrefix(r.URL.Path, authURLPrefix+"/") {
			oidcHandler.ServeHTTP(w, r)
			return
		}

		// If not authenticated, redirect to login and begin the OIDC Authn flow.
		if actr := actor.FromContext(r.Context()); actr == nil || !actr.IsAuthenticated() {
			redirectURL := url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery, Fragment: r.URL.Fragment}
			query := url.Values(map[string][]string{"redirect": []string{redirectURL.String()}})
			http.Redirect(w, r, authURLPrefix+"/login?"+query.Encode(), http.StatusFound)
			return
		}

		// At this point, we've verified that the request has a valid sesssion, so serve the requested resource.
		handler.ServeHTTP(w, r)
	})), nil
}

func startAnonUserSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	token := &oidc.IDToken{
		Issuer:  "oidc-override",
		Subject: "anon-user",
	}
	userInfo := &oidc.UserInfo{
		Subject:       "anon-user",
		Profile:       "anon-user",
		Email:         "anon-user@sourcegraph.com",
		EmailVerified: true,
	}
	claims := &UserClaims{
		Name:              "Anonymous User",
		GivenName:         "Anonymous User",
		FamilyName:        "User",
		PreferredUsername: "anon-user",
	}
	actr, err := getActor(ctx, token, userInfo, claims)
	if err != nil {
		return err
	}
	return session.StartNewSession(w, r, actr, 0)
}

// newOIDCLoginHandler returns a handler that defines the necessary endpoints for the OIDC Authentication Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) on the Relying Party's end.
//
// 🚨 SECURITY
func newOIDCLoginHandler(createCtx context.Context, handler http.Handler, appURL string) (http.Handler, error) {
	provider, err := oidc.NewProvider(createCtx, oidcProvider.Issuer)
	if err != nil {
		return nil, err
	}
	oauth2Config := oauth2.Config{
		ClientID:     oidcProvider.ClientID,
		ClientSecret: oidcProvider.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s%s/%s", appURL, authURLPrefix, "callback"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oidcProvider.ClientID})

	r := mux.NewRouter()

	// Endpoint that starts the Authentication Request Code Flow.
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// The state parameter is an opaque value used to maintain state between the original Authentication Request
		// and the callback. We do not record any state beyond a CSRF token used to defend against CSRF attacks against the callback.
		// We use the CSRF token created by gorilla/csrf that is used for other app endpoints as the OIDC state parameter.
		//
		// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the OIDC spec.
		state := (&authnState{CSRFToken: csrf.Token(r), Redirect: r.URL.Query().Get("redirect")}).Encode()
		http.SetCookie(w, &http.Cookie{Name: oidcStateCookieName, Value: state, Expires: time.Now().Add(time.Minute * 15)})

		// Redirect to the OP's Authorization Endpoint for authentication. The nonce is an optional
		// string value used to associate a Client session with an ID Token and to mitigate replay attacks.
		// Whereas the state parameter is used in validating the Authentication Request
		// callback, the nonce is used in validating the response to the ID Token request.
		// We re-use the Authn request state as the nonce.
		//
		// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the OIDC spec.
		http.Redirect(w, r, oauth2Config.AuthCodeURL(state, oidc.Nonce(state)), http.StatusFound)
	})

	// Endpoint for the OIDC Authorization Response. See http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
	r.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if authError := r.URL.Query().Get("error"); authError != "" {
			http.Error(w, ssoErrMsg(authError, r.URL.Query().Get("error_description")), http.StatusUnauthorized)
			return
		}

		// Validate state parameter to prevent CSRF attacks
		stateParam := r.URL.Query().Get("state")
		if stateParam == "" {
			http.Error(w, ssoErrMsg("No OIDC state query parameter specified", ""), http.StatusBadRequest)
			return
		}
		stateCookie, err := r.Cookie(oidcStateCookieName)
		if err == http.ErrNoCookie {
			http.Error(w, ssoErrMsg("No OIDC state cookie found", "possible request forgery"), http.StatusBadRequest)
			return
		} else if err != nil {
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
			http.Error(w, ssoErrMsg("OIDC state parameter was invalid", ""), http.StatusBadRequest)
			return
		}

		// Exchange the code for an access token. See http://openid.net/specs/openid-connect-core-1_0.html#TokenRequest.
		oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
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
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if oidcProvider.RequireEmailDomain != "" && !strings.HasSuffix(userInfo.Email, "@"+oidcProvider.RequireEmailDomain) {
			http.Error(w, ssoErrMsg("Invalid email domain", "Required: "+oidcProvider.RequireEmailDomain), http.StatusUnauthorized)
			return
		}

		var claims UserClaims
		if err := userInfo.Claims(&claims); err != nil {
			log15.Warn("Could not parse userInfo claims", "error", err)
		}
		actr, err := getActor(ctx, idToken, userInfo, &claims)
		if err != nil {
			log15.Error("Could not get user for OIDC authentication", "error", err)
			http.Error(w, "Could not get user (a user with your email or username may already exist).", http.StatusInternalServerError)
			return
		}
		if err := session.StartNewSession(w, r, actr, 0); err != nil {
			http.Error(w, ssoErrMsg("Could not initiate session", err), http.StatusInternalServerError)
			return
		}

		// Redirect to the page the user was trying to access before login. To prevent an open-redirect vulnerability,
		// strip the host from the redirect URL, so only relative redirects are valid.
		//
		// 🚨 SECURITY
		var redirect string
		if redirectURL, err := url.Parse(state.Redirect); err == nil {
			redirectURL.Scheme = ""
			redirectURL.Host = ""
			redirect = redirectURL.String()
			if !strings.HasPrefix(redirect, "/") {
				redirect = "/" + redirect
			}
		} else {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusFound)
	})
	return http.StripPrefix(authURLPrefix, handlerutil.NewHandlerWithCSRFProtection(r, globals.AppURL.Scheme == "https")), nil
}

// getActor returns the actor corresponding to the user indicated by the OIDC ID Token and UserInfo response.
// Because Actors must correspond to users in our DB, it creates the user in the DB if the user does not yet
// exist.
func getActor(ctx context.Context, idToken *oidc.IDToken, userInfo *oidc.UserInfo, claims *UserClaims) (*actor.Actor, error) {
	provider := idToken.Issuer
	authID := oidcToAuthID(provider, idToken.Subject)
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
	login, err := NormalizeUsername(login)
	if err != nil {
		return nil, err
	}

	usr, err := localstore.Users.GetByAuthID(ctx, authID)
	if _, notFound := err.(localstore.ErrUserNotFound); notFound {
		usr, err = localstore.Users.Create(ctx, authID, email, login, displayName, provider, nil, "", "")
	}
	if err != nil {
		return nil, err
	}

	return actor.FromUser(usr), nil
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

func oidcToAuthID(issuer, subject string) string {
	return fmt.Sprintf("%s:%s", issuer, subject)
}
