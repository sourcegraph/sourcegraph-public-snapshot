package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/csrf"
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
	Picture           string `json:"picture"`
}

// newOIDCAuthMiddleware returns middlewares for OpenID Connect (OIDC) authentication, adding endpoints
// under the auth path prefix ("/.auth") to enable the login flow and requiring login for all other endpoints.
//
// The OIDC spec (http://openid.net/specs/openid-connect-core-1_0.html) describes an authentication protocol
// that involves 3 parties: the Relying Party (e.g., Sourcegraph), the OpenID Provider (e.g., Okta, OneLogin,
// or another SSO provider), and the End User (e.g., a user's web browser).
//
// The middlewares this method returns implement two things: (1) the OIDC Authorization Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) and (2) Sourcegraph-specific session management
// (outside the scope of the OIDC spec). Upon successful completion of the OIDC login flow, the handler will create
// a new session and session cookie. The expiration of the session is the expiration of the OIDC ID Token.
//
// 🚨 SECURITY
func newOIDCAuthMiddleware(createCtx context.Context, appURL string) (*Middleware, error) {
	// Return an error if the OIDC parameters are unset or missing
	if oidcProvider == nil {
		return nil, errors.New("No OpenID Connect Provider specified")
	}
	if oidcProvider.ClientID == "" || oidcProvider.ClientSecret == "" {
		return nil, fmt.Errorf("OIDC Client ID or Client Secret was empty")
	}

	authIfNeededMiddleware := func(nextIfAuthed, nextIfUnauthed http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Respect already authenticated actor (e.g., via access token).
			if actor.FromContext(r.Context()).IsAuthenticated() {
				nextIfAuthed.ServeHTTP(w, r)
				return
			}

			// Otherwise require OpenID authentication.
			nextIfUnauthed.ServeHTTP(w, r)
		})
	}

	// Create handler for OIDC Authentication Code Flow endpoints
	oidcHandler, err := newOIDCLoginHandler(createCtx, appURL)
	if err != nil {
		return nil, err
	}
	return &Middleware{
		API: func(next http.Handler) http.Handler {
			return authIfNeededMiddleware(next, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "requires authentication", http.StatusUnauthorized)
				return
			}))
		},
		App: func(next http.Handler) http.Handler {
			next = authIfNeededMiddleware(next, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Not authenticated: redirect to login and begin the OIDC Authn flow.
				redirectURL := url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery, Fragment: r.URL.Fragment}
				query := url.Values(map[string][]string{"redirect": {redirectURL.String()}})
				http.Redirect(w, r, authURLPrefix+"/login?"+query.Encode(), http.StatusFound)
				return
			}))

			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// If the path is under the authentication path, serve the OIDC Authentication Code Flow handler
				if strings.HasPrefix(r.URL.Path, authURLPrefix+"/") {
					oidcHandler.ServeHTTP(w, r)
					return
				}

				// Otherwise proceed (will check auth).
				next.ServeHTTP(w, r)
			})
		},
	}, nil
}

// newOIDCLoginHandler returns a handler that defines the necessary endpoints for the OIDC Authentication Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) on the Relying Party's end.
//
// 🚨 SECURITY
func newOIDCLoginHandler(createCtx context.Context, appURL string) (http.Handler, error) {
	// Log when fetching the OIDC config from the provider is slow. (It blocks frontend startup.)
	// This can happen on very high latency connections, or when the provider is unreachable.
	timer := time.AfterFunc(5*time.Second, func() {
		log15.Warn("Retrieving OpenID Connect metadata for SSO authentication is taking longer than expected.", "url", oidcProvider.Issuer)
	})
	provider, err := oidc.NewProvider(createCtx, oidcProvider.Issuer)
	timer.Stop()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving OpenID Connect metadata from issuer")
	}
	oauth2Config := oauth2.Config{
		ClientID:     oidcProvider.ClientID,
		ClientSecret: oidcProvider.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s%s/%s", appURL, authURLPrefix, "callback"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oidcProvider.ClientID})

	r := http.NewServeMux()

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
		stateCookie, err := r.Cookie(oidcStateCookieName)
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
			log15.Error("Error looking up OpenID-authenticated user.", "error", err)
			http.Error(w, "Error looking up OpenID-authenticated user. "+couldNotGetUserDescription, http.StatusInternalServerError)
			return
		}
		if err := session.StartNewSession(w, r, actr, 0); err != nil {
			log15.Error("Could not initiate session", "error", err)
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
	return http.StripPrefix(authURLPrefix, r), nil
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
	login, err := NormalizeUsername(login)
	if err != nil {
		return nil, err
	}

	userID, err := createOrUpdateUser(ctx, db.NewUser{
		ExternalProvider: provider,
		ExternalID:       externalID,
		Username:         login,
		Email:            email,
		DisplayName:      displayName,
		AvatarURL:        claims.Picture,
	})
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
