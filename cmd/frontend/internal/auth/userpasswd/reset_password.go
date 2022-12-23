package userpasswd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var defaultResetPasswordEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `Reset your Sourcegraph password ({{.Host}})`,
	Text: `
Somebody (likely you) requested a password reset for the user {{.Username}} on Sourcegraph ({{.Host}}).

To reset the password for {{.Username}} on Sourcegraph, follow this link:

  {{.URL}}
`,
	HTML: `
<p>
  Somebody (likely you) requested a password reset for <strong>{{.Username}}</strong>
  on Sourcegraph ({{.Host}}).
</p>

<p><strong><a href="{{.URL}}">Reset password for {{.Username}}</a></strong></p>
`,
})

func SendResetPasswordURLEmail(ctx context.Context, email, username string, resetURL *url.URL) error {
	// Configure the template
	emailTemplate := defaultResetPasswordEmailTemplates
	if customTemplates := conf.SiteConfig().EmailTemplates; customTemplates != nil {
		emailTemplate = txemail.FromSiteConfigTemplateWithDefault(customTemplates.ResetPassword, emailTemplate)
	}

	return txemail.Send(ctx, "password_reset", txemail.Message{
		To:       []string{email},
		Template: emailTemplate,
		Data: struct {
			Username string
			URL      string
			Host     string
		}{
			Username: username,
			URL:      globals.ExternalURL().ResolveReference(resetURL).String(),
			Host:     globals.ExternalURL().Host,
		},
	})
}

// HandleResetPasswordInit initiates the builtin-auth password reset flow by sending a password-reset email.
func HandleResetPasswordInit(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleResetPasswordInit", "password reset initialization flow handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
			return
		}
		if handleNotAuthenticatedCheck(w, r) {
			return
		}
		if !conf.CanSendEmail() {
			httpLogError(logger.Error, w, "Unable to reset password because email sending is not configured on this site", http.StatusNotFound)
			return
		}

		ctx := r.Context()
		var formData struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
			httpLogError(logger.Error, w, "Could not decode password reset request body", http.StatusBadRequest, log.Error(err))
			return
		}

		if formData.Email == "" {
			httpLogError(logger.Warn, w, "No email specified in password reset request", http.StatusBadRequest)
			return
		}

		usr, err := db.Users().GetByVerifiedEmail(ctx, formData.Email)
		if err != nil {
			// 🚨 SECURITY: We don't show an error message when the user is not found
			// as to not leak the existence of a given e-mail address in the database.
			if !errcode.IsNotFound(err) {
				httpLogError(logger.Warn, w, "Failed to lookup user", http.StatusInternalServerError)
			}
			return
		}

		resetURL, err := backend.MakePasswordResetURL(ctx, db, usr.ID)
		if err == database.ErrPasswordResetRateLimit {
			httpLogError(logger.Warn, w, "Too many password reset requests. Try again in a few minutes.", http.StatusTooManyRequests, log.Error(err))
			return
		} else if err != nil {
			httpLogError(logger.Error, w, "Could not reset password", http.StatusBadRequest, log.Error(err))
			return
		}

		if err := SendResetPasswordURLEmail(r.Context(), formData.Email, usr.Username, resetURL); err != nil {
			httpLogError(logger.Error, w, "Could not send reset password email", http.StatusInternalServerError, log.Error(err))
			return
		}
		database.LogPasswordEvent(ctx, db, r, database.SecurityEventNamPasswordResetRequested, usr.ID)
	}
}

// HandleResetPasswordCode resets the password if the correct code is provided.
func HandleResetPasswordCode(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("HandleResetPasswordCode", "verifies password reset code requests handler")
	return func(w http.ResponseWriter, r *http.Request) {
		if handleEnabledCheck(logger, w) {
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
			httpLogError(logger.Error, w, "Password reset with code: could not decode request body", http.StatusBadGateway, log.Error(err))
			return
		}

		if err := database.CheckPassword(params.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		success, err := db.Users().SetPassword(ctx, params.UserID, params.Code, params.Password)
		if err != nil {
			httpLogError(logger.Error, w, "Unexpected error", http.StatusInternalServerError, log.Error(err))
			return
		}

		if !success {
			http.Error(w, "Password reset code was invalid or expired.", http.StatusUnauthorized)
			return
		}

		database.LogPasswordEvent(ctx, db, r, database.SecurityEventNamePasswordChanged, params.UserID)

		if conf.CanSendEmail() {
			if err := backend.NewUserEmailsService(db, logger).SendUserEmailOnFieldUpdate(ctx, params.UserID, "reset the password"); err != nil {
				logger.Warn("Failed to send email to inform user of password reset", log.Error(err))
			}
		}
	}
}

func handleNotAuthenticatedCheck(w http.ResponseWriter, r *http.Request) (handled bool) {
	if actor.FromContext(r.Context()).IsAuthenticated() {
		http.Error(w, "Authenticated users may not perform password reset.", http.StatusInternalServerError)
		return true
	}
	return false
}
