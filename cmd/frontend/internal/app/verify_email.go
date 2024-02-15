package app

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func serveVerifyEmail(db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := log.Scoped("verify-email")
		email := r.URL.Query().Get("email")
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
		usr, err := db.Users().GetByCurrentAuthUser(ctx)
		if err != nil {
			httpLogAndError(w, logger, "Could not get current user", http.StatusUnauthorized)
			return
		}
		email, alreadyVerified, err := db.UserEmails().Get(ctx, usr.ID, email)
		if err != nil {
			http.Error(w, fmt.Sprintf("No email %q found for user %d", email, usr.ID), http.StatusBadRequest)
			return
		}
		if alreadyVerified {
			http.Error(w, fmt.Sprintf("User %d email %q is already verified", usr.ID, email), http.StatusBadRequest)
			return
		}
		verified, err := db.UserEmails().Verify(ctx, usr.ID, email, verifyCode)
		if err != nil {
			httpLogAndError(w, logger, "Could not verify user email", http.StatusInternalServerError, log.Int32("userID", usr.ID), log.String("email", email), log.Error(err))
			return
		}
		if !verified {
			http.Error(w, "Could not verify user email. Email verification code did not match.", http.StatusUnauthorized)
			return
		}
		// Set the verified email as primary if user has no primary email
		_, _, err = db.UserEmails().GetPrimaryEmail(ctx, usr.ID)
		if err != nil {
			if err := db.UserEmails().SetPrimaryEmail(ctx, usr.ID, email); err != nil {
				httpLogAndError(w, logger, "Could not set primary email.", http.StatusInternalServerError, log.Int32("userID", usr.ID), log.String("email", email), log.Error(err))
				return
			}
		}

		if err := db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameEmailVerified, r.URL.Path, uint32(actr.UID), "", "BACKEND", email); err != nil {
			logger.Warn("Error logging security event", log.Error(err))
		}

		if err = db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
			UserID: usr.ID,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
		}); err != nil {
			logger.Error("Failed to grant user pending permissions", log.Int32("userID", usr.ID), log.Error(err))
		}

		http.Redirect(w, r, "/search", http.StatusFound)
	}
}

func httpLogAndError(w http.ResponseWriter, logger log.Logger, msg string, code int, errArgs ...log.Field) {
	logger.Error(msg, errArgs...)
	http.Error(w, msg, code)
}
