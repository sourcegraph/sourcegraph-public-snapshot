package userpasswd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type credentials struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleSignUp handles submission of the user signup form.
func HandleSignUp(w http.ResponseWriter, r *http.Request) {
	if conf.AuthProvider().Builtin == nil {
		http.Error(w, "Builtin auth provider is not enabled.", http.StatusForbidden)
		return
	}
	if !conf.AuthAllowSignup() {
		http.Error(w, "Signup is not enabled (builtin auth provider allowSignup site configuration option)", http.StatusNotFound)
		return
	}
	handleSignUp(w, r, false)
}

// HandleSiteInit handles submission of the site initialization form, where the initial site admin user is created.
func HandleSiteInit(w http.ResponseWriter, r *http.Request) {
	// This only succeeds if the site is not yet initialized and there are no users yet. It doesn't
	// allow signups after those conditions become true, so we don't need to check auth.allowSignup
	// in site config.
	handleSignUp(w, r, true)
}

// doServeSignUp is called to create a new user account. It is called for the normal user signup process (where a
// non-admin user is created) and for the site initialization process (where the initial site admin user account is
// created).
//
// 🚨 SECURITY: Any change to this function could introduce security exploits
// and/or break sign up / initial admin account creation. Be careful.
func handleSignUp(w http.ResponseWriter, r *http.Request, failIfNewUserIsNotInitialSiteAdmin bool) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	const defaultErrorMessage = "Signup failed unexpectedly."

	// Create the user.
	//
	// We don't need to check auth.allowSignup because we assume the caller of doServeSignUp checks
	// it, or else that failIfNewUserIsNotInitialSiteAdmin == true (in which case the only signup
	// allowed is that of the initial site admin).
	emailCode, err := backend.MakeEmailVerificationCode()
	if err != nil {
		log15.Error("Error generating email verification code for new user.", "email", creds.Email, "username", creds.Username, "error", err)
		http.Error(w, defaultErrorMessage, http.StatusInternalServerError)
		return
	}
	usr, err := db.Users.Create(r.Context(), db.NewUser{
		Email:                 creds.Email,
		Username:              creds.Username,
		Password:              creds.Password,
		EmailVerificationCode: emailCode,
		FailIfNotInitialUser:  failIfNewUserIsNotInitialSiteAdmin,
	})
	if err != nil {
		var (
			message    string
			statusCode int
		)
		switch {
		case db.IsUsernameExists(err):
			message = "Username is already in use. Try a different username."
			statusCode = http.StatusConflict
		case db.IsEmailExists(err):
			message = "Email address is already in use. Try signing into that account instead, or use a different email address."
			statusCode = http.StatusConflict
		default:
			// Do not show non-whitelisted error messages to user, in case they contain sensitive or confusing
			// information.
			message = defaultErrorMessage
			statusCode = http.StatusInternalServerError
		}
		log15.Error("Error in user signup.", "email", creds.Email, "username", creds.Username, "error", err)
		http.Error(w, message, statusCode)
		return
	}
	actor := &actor.Actor{UID: usr.ID}

	if conf.EmailVerificationRequired() {
		if err := backend.SendUserEmailVerificationEmail(r.Context(), creds.Email, emailCode); err != nil {
			log15.Error("failed to send email verification (continuing, user's email will be unverified)", "email", creds.Email, "err", err)
		}
	}

	// Write the session cookie
	if session.SetActor(w, r, actor, 0); err != nil {
		httpLogAndError(w, "Could not create new user session", http.StatusInternalServerError)
	}

	// Track user data
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go tracking.TrackUser(usr.AvatarURL, creds.Email, "SignupCompleted")
	}
}

func getByEmailOrUsername(ctx context.Context, emailOrUsername string) (*types.User, error) {
	if strings.Contains(emailOrUsername, "@") {
		return db.Users.GetByVerifiedEmail(ctx, emailOrUsername)
	}
	return db.Users.GetByUsername(ctx, emailOrUsername)
}

// HandleSignIn accepts a POST containing username-password credentials and authenticates the
// current session if the credentials are valid.
func HandleSignIn(w http.ResponseWriter, r *http.Request) {
	if conf.AuthProvider().Builtin == nil {
		http.Error(w, "Builtin auth provider is not enabled.", http.StatusForbidden)
		return
	}

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

	// Validate user. Allow login by both email and username (for convenience).
	usr, err := getByEmailOrUsername(ctx, creds.Email)
	if err != nil {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized, "err", err)
		return
	}
	// 🚨 SECURITY: check password
	correct, err := db.Users.IsPassword(ctx, usr.ID, creds.Password)
	if err != nil {
		httpLogAndError(w, "Error checking password", http.StatusInternalServerError, "err", err)
		return
	}
	if !correct {
		httpLogAndError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}
	actor := &actor.Actor{UID: usr.ID}

	// Write the session cookie
	if session.SetActor(w, r, actor, 0); err != nil {
		httpLogAndError(w, "Could not create new user session", http.StatusInternalServerError)
		return
	}

	// Track user data
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go tracking.TrackUser(usr.AvatarURL, creds.Email, "SigninCompleted")
	}
}

// HandleResetPasswordInit initiates the builtin-auth password reset flow by sending a password-reset email.
func HandleResetPasswordInit(w http.ResponseWriter, r *http.Request) {
	if conf.AuthProvider().Builtin == nil {
		http.Error(w, "Builtin auth provider is not enabled.", http.StatusForbidden)
		return
	}
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

	usr, err := db.Users.GetByVerifiedEmail(ctx, creds.Email)
	if err != nil {
		httpLogAndError(w, "No user found for email", http.StatusBadRequest, "email", creds.Email)
		return
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, usr.ID, creds.Email)
	if err == db.ErrPasswordResetRateLimit {
		httpLogAndError(w, "Too many password reset requests. Try again in a few minutes.", http.StatusTooManyRequests, "err", err)
		return
	} else if err != nil {
		httpLogAndError(w, "Could not reset password", http.StatusBadRequest, "err", err)
		return
	}

	if err := txemail.Send(r.Context(), txemail.Message{
		To:       []string{creds.Email},
		Template: resetPasswordEmailTemplates,
		Data: struct {
			Username string
			URL      string
		}{
			Username: usr.Username,
			URL:      globals.AppURL.ResolveReference(resetURL).String(),
		},
	}); err != nil {
		httpLogAndError(w, "Could not reset password", http.StatusInternalServerError, "err", err)
		return
	}
}

var (
	resetPasswordEmailTemplates = txemail.MustValidate(txemail.Templates{
		Subject: `Reset your Sourcegraph password`,
		Text: `
Somebody (likely you) requested a password reset for the user {{.Username}} on Sourcegraph.

To reset the password for {{.Username}} on Sourcegraph, follow this link:

  {{.URL}}
`,
		HTML: `
<p>
  Somebody (likely you) requested a password reset for <strong>{{.Username}}</strong>
  on Sourcegraph.
</p>

<p><strong><a href="{{.URL}}">Reset password for {{.Username}}</a></strong></p>
`,
	})
)

// HandleResetPassword resets the password if the correct code is provided.
func HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if conf.AuthProvider().Builtin == nil {
		http.Error(w, "Builtin auth provider is not enabled.", http.StatusForbidden)
		return
	}

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
	usr, err := db.Users.GetByVerifiedEmail(ctx, params.Email)
	if err != nil {
		httpLogAndError(w, fmt.Sprintf("User with email %s not found", params.Email), http.StatusNotFound)
		return
	}

	success, err := db.Users.SetPassword(ctx, usr.ID, params.Code, params.Password)
	if err != nil {
		httpLogAndError(w, "Unexpected error", http.StatusInternalServerError, "err", err)
		return
	}

	if !success {
		httpLogAndError(w, "Password reset failed", http.StatusUnauthorized)
		return
	}
}

func httpLogAndError(w http.ResponseWriter, msg string, code int, errArgs ...interface{}) {
	log15.Error(msg, errArgs...)
	http.Error(w, msg, code)
}
