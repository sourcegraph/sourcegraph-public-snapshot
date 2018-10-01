package userpasswd

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
)

// HandleResetPasswordInit initiates the builtin-auth password reset flow by sending a password-reset email.
func HandleResetPasswordInit(w http.ResponseWriter, r *http.Request) {
	if handleEnabledCheck(w) {
		return
	}
	if handleNotAuthenticatedCheck(w, r) {
		return
	}
	if !conf.CanSendEmail() {
		httpLogAndError(w, "Unable to reset password because email sending is not configured on this site", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	var formData struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		httpLogAndError(w, "Could not decode password reset request body", http.StatusBadRequest, "err", err)
		return
	}

	if formData.Email == "" {
		httpLogAndError(w, "No email specified in password reset request", http.StatusBadRequest)
		return
	}

	usr, err := db.Users.GetByVerifiedEmail(ctx, formData.Email)
	if err != nil {
		httpLogAndError(w, "No user found with a matching verified email address", http.StatusBadRequest, "email", formData.Email)
		return
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, usr.ID)
	if err == db.ErrPasswordResetRateLimit {
		httpLogAndError(w, "Too many password reset requests. Try again in a few minutes.", http.StatusTooManyRequests, "err", err)
		return
	} else if err != nil {
		httpLogAndError(w, "Could not reset password", http.StatusBadRequest, "err", err)
		return
	}

	if err := txemail.Send(r.Context(), txemail.Message{
		To:       []string{formData.Email},
		Template: resetPasswordEmailTemplates,
		Data: struct {
			Username string
			URL      string
		}{
			Username: usr.Username,
			URL:      globals.AppURL.ResolveReference(resetURL).String(),
		},
	}); err != nil {
		httpLogAndError(w, "Could not send reset password email", http.StatusInternalServerError, "err", err)
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

// HandleResetPasswordCode resets the password if the correct code is provided.
func HandleResetPasswordCode(w http.ResponseWriter, r *http.Request) {
	if handleEnabledCheck(w) {
		return
	}
	if handleNotAuthenticatedCheck(w, r) {
		return
	}

	ctx := r.Context()
	var params struct {
		UserID   int32  `json:"userID"`
		Code     string `json:"code"`
		Password string `json:"password"` // new password
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httpLogAndError(w, "Password reset with code: could not decode request body", http.StatusBadGateway, "err", err)
		return
	}

	success, err := db.Users.SetPassword(ctx, params.UserID, params.Code, params.Password)
	if err != nil {
		httpLogAndError(w, "Unexpected error", http.StatusInternalServerError, "err", err)
		return
	}

	if !success {
		httpLogAndError(w, "Password reset failed", http.StatusUnauthorized)
		return
	}
}

func handleNotAuthenticatedCheck(w http.ResponseWriter, r *http.Request) (handled bool) {
	if actor.FromContext(r.Context()).IsAuthenticated() {
		http.Error(w, "Authenticated users may not perform password reset.", http.StatusInternalServerError)
		return true
	}
	return false
}
