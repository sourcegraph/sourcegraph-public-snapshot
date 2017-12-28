package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	"github.com/mattbaird/gochimp"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type oauthCookie struct {
	Nonce       string
	RedirectURL string
	ReturnTo    string
	ReturnToNew string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// serveSignUp serves the native-auth sign-up endpoint
func serveSignUp(w http.ResponseWriter, r *http.Request) {
	if !conf.Get().AuthAllowSignup {
		http.Error(w, "signup is not enabled", http.StatusNotFound)
	}
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	displayName := creds.DisplayName
	if displayName == "" {
		displayName = creds.Username
	}

	// Create user
	emailCode := backend.MakeEmailVerificationCode()
	usr, err := store.Users.Create(r.Context(), backend.NativeAuthUserAuthID(creds.Email), creds.Email, creds.Username, displayName, sourcegraph.UserProviderNative, nil, creds.Password, emailCode)
	if err != nil {
		httpLogAndError(w, fmt.Sprintf("Could not create user %s", creds.Username), http.StatusInternalServerError)
		return
	}
	actor := &actor.Actor{
		UID:   usr.Auth0ID,
		Login: usr.Username,
		Email: usr.Email,
	}

	if conf.EmailVerificationRequired() {
		// Send verify email
		q := make(url.Values)
		q.Set("code", emailCode)
		verifyLink := globals.AppURL.String() + router.Rel.URLTo(router.VerifyEmail).Path + "?" + q.Encode()
		notif.SendMandrillTemplate(&notif.EmailConfig{
			Template:  "verify-email",
			FromEmail: "noreply@sourcegraph.com",
			ToEmail:   creds.Email,
			Subject:   "Verify your email on Sourcegraph Server",
		}, []gochimp.Var{}, []gochimp.Var{{Name: "VERIFY_URL", Content: verifyLink}})
	}

	// Write the session cookie
	if session.StartNewSession(w, r, actor, 0); err != nil {
		httpLogAndError(w, "Could not create new user session", http.StatusInternalServerError)
		return
	}
}

func getUserFromNativeOrAuth0(ctx context.Context, email string) (*sourcegraph.User, error) {
	authID := backend.NativeAuthUserAuthID(email)
	usr, err := store.Users.GetByAuth0ID(ctx, authID)
	if err == nil {
		return usr, nil
	} else if err != nil {
		if _, ok := err.(store.ErrUserNotFound); !ok {
			return nil, err
		}
	}

	// The user might be a legacy auth0 user.
	return store.Users.GetByEmail(ctx, email)
}

// serveSignIn2 serves a native-auth endpoint
func serveSignIn2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("Unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Could not decode request body", http.StatusBadRequest)
		return
	}

	// Validate user
	usr, err := getUserFromNativeOrAuth0(ctx, creds.Email)
	if err != nil {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", err)
		return
	}
	if conf.AuthProvider() != "auth0" && usr.Provider != sourcegraph.UserProviderNative {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", "not a native auth user")
		return
	}
	// 🚨 SECURITY: check password
	correct, err := store.Users.IsPassword(ctx, usr.ID, creds.Password)
	if err != nil {
		httpLogAndError(w, "Error checking password", http.StatusInternalServerError, "err", err)
		return
	}
	if !correct {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}
	actor := &actor.Actor{
		UID:   usr.Auth0ID,
		Login: usr.Username,
		Email: usr.Email,
	}

	// Write the session cookie
	if session.StartNewSession(w, r, actor, 0); err != nil {
		httpLogAndError(w, "Could not create new user session", http.StatusInternalServerError)
		return
	}

	// Track user data in GCS
	eventLabel := "CompletedNativeSignIn"
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go tracking.TrackUser(actor, eventLabel)
	}
}

// serveVerifyEmail serves the email verification link for native auth
func serveVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	verifyCode := r.URL.Query().Get("code")
	actr := actor.FromContext(ctx)
	if !actr.IsAuthenticated() {
		redirectTo := r.URL.String()
		q := make(url.Values)
		q.Set("returnTo", redirectTo)
		http.Redirect(w, r, "/sign-in?"+q.Encode(), http.StatusFound)
		return
	}
	// 🚨 SECURITY: require correct authed user to verify email
	usr, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		httpLogAndError(w, "Could not get current user", http.StatusUnauthorized)
		return
	}
	if conf.AuthProvider() != "auth0" && usr.Provider != sourcegraph.UserProviderNative {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", "not a native auth user")
		return
	}
	if usr.Verified {
		http.Error(w, fmt.Sprintf("User %s already verified", usr.Email), http.StatusBadRequest)
		return
	}
	verified, err := store.Users.ValidateEmail(ctx, usr.ID, verifyCode)
	if err != nil {
		http.Error(w, "Unexpected error when verifying user.", http.StatusInternalServerError)
		return
	}
	if !verified {
		http.Error(w, "Could not verify user. Code did not match email.", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusFound)
}

// serveResetPasswordInit initiates the native-auth password reset flow by sending a password-reset email.
func serveResetPasswordInit(w http.ResponseWriter, r *http.Request) {
	if !conf.CanSendEmail() {
		httpLogAndError(w, "Unable to reset password because email sending is not configured on this site", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		httpLogAndError(w, "Could not decode password reset request body", http.StatusBadRequest, "err", err)
		return
	}

	if creds.Email == "" {
		httpLogAndError(w, "No email specified in password reset request", http.StatusBadRequest)
		return
	}

	usr, err := store.Users.GetByEmail(ctx, creds.Email)
	if err != nil {
		httpLogAndError(w, "No user found for email", http.StatusBadRequest, "email", creds.Email)
		return
	}
	if conf.AuthProvider() != "auth0" && usr.Provider != sourcegraph.UserProviderNative {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", "not a native auth user")
		return
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, usr.ID, creds.Email)
	if err == localstore.ErrPasswordResetRateLimit {
		httpLogAndError(w, "Too many password reset requests. Try again in a few minutes.", http.StatusTooManyRequests, "err", err)
		return
	} else if err != nil {
		httpLogAndError(w, "Could not reset password", http.StatusBadRequest, "err", err)
		return
	}

	notif.SendMandrillTemplate(&notif.EmailConfig{
		Template:  "forgot-password",
		FromEmail: "noreply@sourcegraph.com",
		ToEmail:   usr.Email,
		Subject:   "Reset your Sourcegraph Server password",
	}, []gochimp.Var{}, []gochimp.Var{
		{Name: "SUBJECT", Content: "Reset password"},
		{Name: "LOGIN", Content: usr.Username},
		{Name: "RESET_LINK", Content: globals.AppURL.ResolveReference(resetURL).String()},
	})
}

// serveResetPassword resets a native-auth password if the correct code is provided
func serveResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var params struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Password string `json:"password"` // new password
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httpLogAndError(w, "Password reset with code: could not decode request body", http.StatusBadGateway, "err", err)
		return
	}

	// 🚨 SECURITY: require correct authed user to reset password
	usr, err := store.Users.GetByEmail(ctx, params.Email)
	if err != nil {
		httpLogAndError(w, fmt.Sprintf("User with email %s not found", params.Email), http.StatusNotFound)
		return
	}
	if conf.AuthProvider() != "auth0" && usr.Provider != sourcegraph.UserProviderNative {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", "not a native auth user")
		return
	}

	success, err := store.Users.SetPassword(ctx, usr.ID, backend.NativeAuthUserAuthID(params.Email), params.Code, params.Password)
	if err != nil {
		httpLogAndError(w, "Unexpected error", http.StatusInternalServerError, "err", err)
		return
	}

	if !success {
		httpLogAndError(w, "Password reset failed", http.StatusUnauthorized)
		return
	}
}

func addEditorBetaTag(ctx context.Context, user *sourcegraph.User, tokenString string) (*sourcegraph.UserTag, error) {
	// 🚨 SECURITY: verify that the token is valid before adding editor-beta tag
	_, err := invite.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return store.UserTags.CreateIfNotExists(ctx, user.ID, "editor-beta")
}

func httpLogAndError(w http.ResponseWriter, msg string, code int, errArgs ...interface{}) {
	log15.Error(msg, errArgs...)
	http.Error(w, msg, code)
}
