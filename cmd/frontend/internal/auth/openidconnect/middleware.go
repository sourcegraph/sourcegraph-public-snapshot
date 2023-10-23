// Package openidconnect implements auth via OIDC.
package openidconnect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/csrf"
	"github.com/inconshreveable/log15"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const stateCookieName = "sg-oidc-state"

// All OpenID Connect endpoints are under this path prefix.
const authPrefix = auth.AuthURLPrefix + "/openidconnect"

type userClaims struct {
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
	EmailVerified     *bool  `json:"email_verified"`
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
func Middleware(db database.DB) *auth.Middleware {
	return &auth.Middleware{
		API: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleOpenIDConnectAuth(db, w, r, next, true)
			})
		},
		App: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handleOpenIDConnectAuth(db, w, r, next, false)
			})
		},
	}
}

// handleOpenIDConnectAuth performs OpenID Connect authentication (if configured) for HTTP requests,
// both API requests and non-API requests.
func handleOpenIDConnectAuth(db database.DB, w http.ResponseWriter, r *http.Request, next http.Handler, isAPIRequest bool) {
	// Fixup URL path. We use "/.auth/callback" as the redirect URI for OpenID Connect, but the rest
	// of this middleware's handlers expect paths of "/.auth/openidconnect/...", so add the
	// "openidconnect" path component. We can't change the redirect URI because it is hardcoded in
	// instances' external auth providers.
	if r.URL.Path == auth.AuthURLPrefix+"/callback" {
		// Rewrite "/.auth/callback" -> "/.auth/openidconnect/callback".
		r.URL.Path = authPrefix + "/callback"
	}

	// Delegate to the OpenID Connect auth handler.
	if !isAPIRequest && strings.HasPrefix(r.URL.Path, authPrefix+"/") {
		authHandler(db)(w, r)
		return
	}

	// If the actor is authenticated and not performing an OpenID Connect flow, then proceed to
	// next.
	if sgactor.FromContext(r.Context()).IsAuthenticated() {
		next.ServeHTTP(w, r)
		return
	}

	// If there is only one auth provider configured, the single auth provider is OpenID Connect,
	// it's an app request, and the sign-out cookie is not present, redirect to sign-in immediately.
	//
	// For sign-out requests (sign-out cookie is  present), the user is redirected to the Sourcegraph login page.
	ps := providers.Providers()
	openIDConnectEnabled := len(ps) == 1 && ps[0].Config().Openidconnect != nil
	if openIDConnectEnabled && !auth.HasSignOutCookie(r) && !isAPIRequest {
		p, safeErrMsg, err := GetProviderAndRefresh(r.Context(), ps[0].ConfigID().ID, GetProvider)
		if err != nil {
			log15.Error("Failed to get provider", "error", err)
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			return
		}
		RedirectToAuthRequest(w, r, p, stateCookieName, auth.SafeRedirectURL(r.URL.String()))
		return
	}

	next.ServeHTTP(w, r)
}

// MockVerifyIDToken mocks the OIDC ID Token verification step. It should only be
// set in tests.
var MockVerifyIDToken func(rawIDToken string) *oidc.IDToken

// authHandler handles the OIDC Authentication Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) on the Relying Party's end.
//
// 🚨 SECURITY
func authHandler(db database.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Path, authPrefix) {
		case "/login": // Endpoint that starts the Authentication Request Code Flow.
			p, safeErrMsg, err := GetProviderAndRefresh(r.Context(), r.URL.Query().Get("pc"), GetProvider)
			if err != nil {
				log15.Error("Failed to get provider.", "error", err)
				http.Error(w, safeErrMsg, http.StatusInternalServerError)
				return
			}
			redirect := r.URL.Query().Get("redirect")
			if redirect == "" {
				redirect = r.URL.Query().Get("returnTo")
			}
			RedirectToAuthRequest(w, r, p, stateCookieName, redirect)
			return

		case "/callback": // Endpoint for the OIDC Authorization Response, see http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
			result, safeErrMsg, errStatus, err := AuthCallback(db, r, stateCookieName, "", GetProvider)
			if err != nil {
				log15.Error("Failed to authenticate with OpenID connect.", "error", err)
				http.Error(w, safeErrMsg, errStatus)
				arg, _ := json.Marshal(struct {
					SafeErrorMsg string `json:"safe_error_msg"`
				}{
					SafeErrorMsg: safeErrMsg,
				})
				db.SecurityEventLogs().LogEvent(r.Context(), &database.SecurityEvent{
					Name:            database.SecurityEventOIDCLoginFailed,
					URL:             r.URL.Path,                                   // Don't log OIDC query params
					AnonymousUserID: fmt.Sprintf("unknown OIDC @ %s", time.Now()), // we don't have a reliable user identifier at the time of the failure
					Source:          "BACKEND",
					Timestamp:       time.Now(),
					Argument:        arg,
				})
				return
			}
			db.SecurityEventLogs().LogEvent(r.Context(), &database.SecurityEvent{
				Name:      database.SecurityEventOIDCLoginSucceeded,
				URL:       r.URL.Path, // Don't log OIDC query params
				UserID:    uint32(result.User.ID),
				Source:    "BACKEND",
				Timestamp: time.Now(),
			})

			var exp time.Duration
			// 🚨 SECURITY: TODO(sqs): We *should* uncomment the lines below to make our own sessions
			// only last for as long as the OP said the access token is active for. Unfortunately,
			// until we support refreshing access tokens in the background
			// (https://github.com/sourcegraph/sourcegraph/issues/11340), this provides a bad user
			// experience because users need to re-authenticate via OIDC every minute or so
			// (assuming their OIDC OP, like many, has a 1-minute access token validity period).
			//
			// if !idToken.Expiry.IsZero() {
			// 	exp = time.Until(idToken.Expiry)
			// }
			if err = session.SetActor(w, r, sgactor.FromUser(result.User.ID), exp, result.User.CreatedAt); err != nil {
				log15.Error("Failed to authenticate with OpenID connect: could not initiate session.", "error", err)
				http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not initiate session.", http.StatusInternalServerError)
				return
			}

			if err = session.SetData(w, r, SessionKey, result.SessionData); err != nil {
				// It's not fatal if this fails. It just means we won't be able to sign the user
				// out of the OP.
				log15.Warn("Failed to set OpenID Connect session data. The session is still secure, but Sourcegraph will be unable to revoke the user's token or redirect the user to the end-session endpoint after the user signs out of Sourcegraph.", "error", err)
			}

			// 🚨 SECURITY: Call auth.SafeRedirectURL to avoid the open-redirect vulnerability.
			http.Redirect(w, r, auth.SafeRedirectURL(result.Redirect), http.StatusFound)

		default:
			http.Error(w, "", http.StatusNotFound)
		}
	}
}

// AuthCallbackResult is the result of handling the authentication callback.
type AuthCallbackResult struct {
	User        *types.User // The user that is upserted and authenticated.
	SessionData SessionData // The corresponding session data to be set for the authenticated user.
	Redirect    string      // The redirect URL for the authenticated user.
}

// AuthCallback handles the callback in the authentication flow which validates
// state and upserts the user and returns the result.
//
// In case of an error, it returns the internal error, an error message that is
// safe to be passed back to the user, and a proper HTTP status code
// corresponding to the error.
func AuthCallback(db database.DB, r *http.Request, stateCookieName, usernamePrefix string, getProvider func(id string) *Provider) (result *AuthCallbackResult, safeErrMsg string, errStatus int, err error) {
	if authError := r.URL.Query().Get("error"); authError != "" {
		errorDesc := r.URL.Query().Get("error_description")
		return nil,
			fmt.Sprintf("Authentication failed. Try signing in again (and clearing cookies for the current site). The authentication provider reported the following problems.\n\n%s\n\n%s", authError, errorDesc),
			http.StatusUnauthorized,
			errors.Errorf("%s - %s", authError, errorDesc)
	}

	// Validate state parameter to prevent CSRF attacks
	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		desc := "Authentication failed. Try signing in again (and clearing cookies for the current site). No OpenID Connect state query parameter specified."
		return nil,
			desc,
			http.StatusBadRequest,
			errors.New(desc)
	}

	stateCookie, err := r.Cookie(stateCookieName)
	if err == http.ErrNoCookie {
		return nil,
			fmt.Sprintf("Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: no state cookie found (possible request forgery, or more than %s elapsed since you started the authentication process).", stateCookieTimeout),
			http.StatusBadRequest,
			errors.New("no state cookie found (possible request forgery).")
	} else if err != nil {
		return nil,
			"Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: invalid state cookie.",
			http.StatusInternalServerError,
			errors.Wrap(err, "could not read state cookie (possible request forgery)")
	}
	if stateCookie.Value != stateParam {
		return nil,
			"Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: state parameter did not match the expected value (possible request forgery).",
			http.StatusBadRequest,
			errors.New("state cookie mismatch (possible request forgery)")
	}

	// Decode state param value
	var state AuthnState
	if err = state.Decode(stateParam); err != nil {
		return nil,
			"Authentication failed. OpenID Connect state parameter was malformed.",
			http.StatusBadRequest,
			errors.Wrap(err, "state parameter was malformed")
	}

	p, safeErrMsg, err := GetProviderAndRefresh(r.Context(), state.ProviderID, getProvider)
	if err != nil {
		return nil,
			safeErrMsg,
			http.StatusInternalServerError,
			errors.Wrap(err, "get provider")
	}

	// Exchange the code for an access token, see http://openid.net/specs/openid-connect-core-1_0.html#TokenRequest.
	oauth2Token, err := p.oauth2Config().Exchange(context.WithValue(r.Context(), oauth2.HTTPClient, httpcli.ExternalClient), r.URL.Query().Get("code"))
	if err != nil {
		return nil,
			"Authentication failed. Try signing in again. The error was: unable to obtain access token from issuer.",
			http.StatusUnauthorized,
			errors.Wrap(err, "obtain access token from OP")
	}

	// Extract the ID Token from the Access Token, see http://openid.net/specs/openid-connect-core-1_0.html#TokenResponse.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil,
			"Authentication failed. Try signing in again. The error was: the issuer's authorization response did not contain an ID token.",
			http.StatusUnauthorized,
			errors.New("the issuer's authorization response did not contain an ID token")
	}

	// Parse and verify ID Token payload, see http://openid.net/specs/openid-connect-core-1_0.html#TokenResponseValidation.
	var idToken *oidc.IDToken
	if MockVerifyIDToken != nil {
		idToken = MockVerifyIDToken(rawIDToken)
	} else {
		idToken, err = p.oidcVerifier().Verify(r.Context(), rawIDToken)
		if err != nil {
			return nil,
				"Authentication failed. Try signing in again. The error was: OpenID Connect ID token could not be verified.",
				http.StatusUnauthorized,
				errors.Wrap(err, "verify ID token")
		}
	}

	// Validate the nonce. The Verify method explicitly doesn't handle nonce
	// validation, so we do that here. We set the nonce to be the same as the state
	// in the Authentication Request state, so we check for equality here.
	if idToken.Nonce != stateParam {
		return nil,
			"Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: OpenID Connect nonce is incorrect (possible replay attack).",
			http.StatusUnauthorized,
			errors.New("nonce is incorrect (possible replay attach)")
	}

	userInfo, err := p.oidcUserInfo(oidc.ClientContext(r.Context(), httpcli.ExternalClient), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		return nil,
			"Failed to get userinfo: " + err.Error(),
			http.StatusInternalServerError,
			errors.Wrap(err, "get user info")
	}

	if p.config.RequireEmailDomain != "" && !strings.HasSuffix(userInfo.Email, "@"+p.config.RequireEmailDomain) {
		return nil,
			fmt.Sprintf("Authentication failed. Only users in %q are allowed.", p.config.RequireEmailDomain),
			http.StatusUnauthorized,
			errors.Errorf("user's email %q is not from allowed domain %q", userInfo.Email, p.config.RequireEmailDomain)
	}

	var claims userClaims
	if err = userInfo.Claims(&claims); err != nil {
		log15.Warn("OpenID Connect auth: could not parse userInfo claims.", "error", err)
	}

	getCookie := func(name string) string {
		c, err := r.Cookie(name)
		if err != nil {
			return ""
		}
		return c.Value
	}
	anonymousId, _ := cookie.AnonymousUID(r)
	newUserCreated, actor, safeErrMsg, err := getOrCreateUser(r.Context(), db, p, idToken, userInfo, &claims, usernamePrefix, anonymousId, getCookie("sourcegraphSourceUrl"), getCookie("sourcegraphRecentSourceUrl"))
	if err != nil {
		return nil,
			safeErrMsg,
			http.StatusInternalServerError,
			errors.Wrap(err, "look up authenticated user")
	}

	// Add a ?signup= or ?signin= parameter to the redirect URL.
	redirectURL := auth.AddPostAuthRedirectParametersToString(state.Redirect, newUserCreated, "OpenIDConnect")

	user, err := db.Users().GetByID(r.Context(), actor.UID)
	if err != nil {
		return nil,
			"Failed to retrieve user from database",
			http.StatusInternalServerError,
			errors.Wrap(err, "get user by ID")
	}
	return &AuthCallbackResult{
		User: user,
		SessionData: SessionData{
			ID:          p.ConfigID(),
			AccessToken: oauth2Token.AccessToken,
			TokenType:   oauth2Token.TokenType,
		},
		Redirect: auth.SafeRedirectURL(redirectURL),
	}, "", 0, nil
}

// AuthnState is the state parameter passed to the authentication request and
// returned to the authentication response callback.
type AuthnState struct {
	CSRFToken string `json:"csrfToken"`
	Redirect  string `json:"redirect"`

	// Allow /.auth/callback to demux callbacks from multiple OpenID Connect OPs.
	ProviderID string `json:"p"`
}

// Encode returns the base64-encoded JSON representation of the authn state.
func (s *AuthnState) Encode() string {
	b, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the base64-encoded JSON representation of the authn state into
// the receiver.
func (s *AuthnState) Decode(encoded string) error {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, s)
}

// stateCookieTimeout defines how long the state cookie should be valid for a
// single authentication flow.
const stateCookieTimeout = time.Minute * 15

// RedirectToAuthRequest redirects the user to the authentication endpoint on the
// external authentication provider.
func RedirectToAuthRequest(w http.ResponseWriter, r *http.Request, p *Provider, cookieName, returnToURL string) {
	// The state parameter is an opaque value used to maintain state between the
	// original Authentication Request and the callback. We do not record any state
	// beyond a CSRF token used to defend against CSRF attacks against the callback.
	// We use the CSRF token created by gorilla/csrf that is used for other app
	// endpoints as the OIDC state parameter.
	//
	// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the
	// OIDC spec.
	state := (&AuthnState{
		CSRFToken:  csrf.Token(r),
		Redirect:   returnToURL,
		ProviderID: p.ConfigID().ID,
	}).Encode()
	http.SetCookie(w,
		&http.Cookie{
			Name:    cookieName,
			Value:   state,
			Path:    auth.AuthURLPrefix + "/", // Include the OIDC redirect URI (/.auth/callback not /.auth/openidconnect/callback for backwards compatibility)
			Expires: time.Now().Add(stateCookieTimeout),
		},
	)

	// Redirect to the OP's Authorization Endpoint for authentication. The nonce is
	// an optional string value used to associate a Client session with an ID Token
	// and to mitigate replay attacks. Whereas the state parameter is used in
	// validating the Authentication Request callback, the nonce is used in
	// validating the response to the ID Token request. We re-use the Authn request
	// state as the nonce.
	//
	// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the
	// OIDC spec.
	http.Redirect(w, r, p.oauth2Config().AuthCodeURL(state, oidc.Nonce(state)), http.StatusFound)
}
