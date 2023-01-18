package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UserEmailsService contains backend methods related to user email addresses.
type UserEmailsService interface {
	Add(ctx context.Context, userID int32, email string) error
	Remove(ctx context.Context, userID int32, email string) error
	SetPrimaryEmail(ctx context.Context, userID int32, email string) error
	SetVerified(ctx context.Context, userID int32, email string, verified bool) error
	ResendVerificationEmail(ctx context.Context, userID int32, email string, now time.Time) error
	SendUserEmailOnFieldUpdate(ctx context.Context, id int32, change string) error
}

// NewUserEmailsService creates an instance of UserEmailsService that contains
// backend methods related to user email addresses.
func NewUserEmailsService(db database.DB, logger log.Logger) UserEmailsService {
	return &userEmails{
		db:     db,
		logger: logger,
	}
}

type userEmails struct {
	db     database.DB
	logger log.Logger
}

// Add adds an email address to a user. If email verification is required, it sends an email
// verification email.
func (e *userEmails) Add(ctx context.Context, userID int32, email string) error {
	logger := e.logger.Scoped("UserEmails.Add", "handles addition of user emails")
	// 🚨 SECURITY: Only the user and site admins can add an email address to a user.
	if err := auth.CheckSiteAdminOrSameUser(ctx, e.db, userID); err != nil {
		return err
	}

	// Prevent abuse (users adding emails of other people whom they want to annoy) with the
	// following abuse prevention checks.
	if isSiteAdmin := auth.CheckCurrentUserIsSiteAdmin(ctx, e.db) == nil; !isSiteAdmin {
		abused, reason, err := checkEmailAbuse(ctx, e.db, userID)
		if err != nil {
			return err
		} else if abused {
			return errors.Errorf("refusing to add email address because %s", reason)
		}
	}

	var code *string
	if conf.EmailVerificationRequired() {
		tmp, err := MakeEmailVerificationCode()
		if err != nil {
			return err
		}
		code = &tmp
	}

	// Another user may have already verified this email address. If so, do not send another
	// verification email (it would be pointless and also be an abuse vector). Do not tell the
	// user that another user has already verified it, to avoid needlessly leaking the existence
	// of emails.
	var emailAlreadyExistsAndIsVerified bool
	if _, err := e.db.Users().GetByVerifiedEmail(ctx, email); err != nil && !errcode.IsNotFound(err) {
		return err
	} else if err == nil {
		emailAlreadyExistsAndIsVerified = true
	}

	if err := e.db.UserEmails().Add(ctx, userID, email, code); err != nil {
		return err
	}

	if conf.EmailVerificationRequired() && !emailAlreadyExistsAndIsVerified {
		usr, err := e.db.Users().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		defer func() {
			// Note: We want to mark as sent regardless because every part of the codebase
			// assumed the email sending would never fail and uses the value of the
			// "last_verification_sent_at" column to calculate cooldown (instead of using
			// cache), while still aligning the semantics to the column name.
			if err = e.db.UserEmails().SetLastVerification(ctx, userID, email, *code, time.Now()); err != nil {
				logger.Warn("Failed to set last verification sent at for the user email", log.Int32("userID", userID), log.Error(err))
			}
		}()

		// Send email verification email.
		if err := SendUserEmailVerificationEmail(ctx, usr.Username, email, *code); err != nil {
			return errors.Wrap(err, "SendUserEmailVerificationEmail")
		}
	}
	return nil
}

// Remove removes the e-mail from the specified user. Perforce external accounts
// using the e-mail will also be removed.
func (e *userEmails) Remove(ctx context.Context, userID int32, email string) (err error) {
	logger := e.logger.Scoped("UserEmails.Remove", "handles removal of user emails")

	// 🚨 SECURITY: Only the authenticated user and site admins can remove email
	// from users' accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, e.db, userID); err != nil {
		return err
	}

	tx, err := e.db.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer func() {
		err = tx.Done(err)
		if err != nil {
			return
		}

		// Eagerly attempt to sync permissions again. This needs to happen _after_ the
		// transaction has committed so that it takes into account any changes triggered
		// by the removal of the e-mail.
		triggerPermissionsSync(ctx, logger, e.db, userID, permssync.ReasonUserEmailRemoved)
	}()

	if err := tx.UserEmails().Remove(ctx, userID, email); err != nil {
		return errors.Wrap(err, "removing user e-mail")
	}

	// 🚨 SECURITY: If an email is removed, invalidate any existing password reset
	// tokens that may have been sent to that email.
	if err := tx.Users().DeletePasswordResetCode(ctx, userID); err != nil {
		return errors.Wrap(err, "deleting reset codes")
	}

	if err := deleteStalePerforceExternalAccounts(ctx, tx, userID, email); err != nil {
		return errors.Wrap(err, "removing stale perforce external account")
	}

	if conf.CanSendEmail() {
		svc := NewUserEmailsService(tx, logger)
		if err := svc.SendUserEmailOnFieldUpdate(ctx, userID, "removed an email"); err != nil {
			logger.Warn("Failed to send email to inform user of email removal", log.Error(err))
		}
	}

	return nil
}

// SetPrimaryEmail sets the supplied e-mail address as the primary address for
// the given user.
func (e *userEmails) SetPrimaryEmail(ctx context.Context, userID int32, email string) error {
	logger := e.logger.Scoped("UserEmails.SetPrimaryEmail", "handles setting primary e-mail for user")

	// 🚨 SECURITY: Only the authenticated user and site admins can set the primary
	// email for users' accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, e.db, userID); err != nil {
		return err
	}

	if err := e.db.UserEmails().SetPrimaryEmail(ctx, userID, email); err != nil {
		return err
	}

	if conf.CanSendEmail() {
		if err := e.SendUserEmailOnFieldUpdate(ctx, userID, "changed primary email"); err != nil {
			logger.Warn("Failed to send email to inform user of primary address change", log.Error(err))
		}
	}

	return nil
}

// SetVerified sets the supplied e-mail as the verified email for the given user.
// If verified is false, Perforce external accounts using the e-mail will be
// removed.
func (e *userEmails) SetVerified(ctx context.Context, userID int32, email string, verified bool) (err error) {
	logger := e.logger.Scoped("UserEmails.SetVerified", "handles setting e-mail as verified")

	// 🚨 SECURITY: Only site admins (NOT users themselves) can manually set email
	// verification status. Users themselves must go through the normal email
	// verification process.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, e.db); err != nil {
		return err
	}

	tx, err := e.db.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer func() {
		err = tx.Done(err)
		if err != nil {
			return
		}

		// Eagerly attempt to sync permissions again. This needs to happen _after_ the
		// transaction has committed so that it takes into account any changes triggered
		// by changes in the verification status of the e-mail.
		triggerPermissionsSync(ctx, logger, e.db, userID, permssync.ReasonUserEmailVerified)
	}()

	if err := tx.UserEmails().SetVerified(ctx, userID, email, verified); err != nil {
		return err
	}

	if !verified {
		if err := deleteStalePerforceExternalAccounts(ctx, tx, userID, email); err != nil {
			return errors.Wrap(err, "removing stale perforce external account")
		}
		return nil
	}

	if err := tx.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
		UserID: userID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}); err != nil {
		logger.Error("schemaResolver.SetUserEmailVerified: failed to grant user pending permissions", log.Int32("userID", userID), log.Error(err))
	}

	return nil
}

// ResendVerificationEmail attempts to re-send the verification e-mail for the
// given user and email combination. If an e-mail sent within the last minute we
// do nothing.
func (e *userEmails) ResendVerificationEmail(ctx context.Context, userID int32, email string, now time.Time) error {
	// 🚨 SECURITY: Only the authenticated user and site admins can resend
	// verification email for their accounts.
	if err := auth.CheckSiteAdminOrSameUser(ctx, e.db, userID); err != nil {
		return err
	}

	user, err := e.db.Users().GetByID(ctx, userID)
	if err != nil {
		return err
	}

	userEmails := e.db.UserEmails()
	lastSent, err := userEmails.GetLatestVerificationSentEmail(ctx, email)
	if err != nil && !errcode.IsNotFound(err) {
		return err
	}
	if lastSent != nil &&
		lastSent.LastVerificationSentAt != nil &&
		now.Sub(*lastSent.LastVerificationSentAt) < 1*time.Minute {
		return errors.New("Last verification email sent too recently")
	}

	email, verified, err := userEmails.Get(ctx, userID, email)
	if err != nil {
		return err
	}
	if verified {
		return nil
	}

	code, err := MakeEmailVerificationCode()
	if err != nil {
		return err
	}

	err = userEmails.SetLastVerification(ctx, userID, email, code, now)
	if err != nil {
		return err
	}

	return SendUserEmailVerificationEmail(ctx, user.Username, email, code)
}

// SendUserEmailOnFieldUpdate sends the user an email that important account information has changed.
// The change is the information we want to provide the user about the change
func (e *userEmails) SendUserEmailOnFieldUpdate(ctx context.Context, id int32, change string) error {
	email, verified, err := e.db.UserEmails().GetPrimaryEmail(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user primary email")
	}
	if !verified {
		return errors.Newf("unable to send email to user ID %d's unverified primary email address", id)
	}
	usr, err := e.db.Users().GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "get user")
	}

	return txemail.Send(ctx, "user_account_update", txemail.Message{
		To:       []string{email},
		Template: updateAccountEmailTemplate,
		Data: struct {
			Email    string
			Change   string
			Username string
			Host     string
		}{
			Email:    email,
			Change:   change,
			Username: usr.Username,
			Host:     globals.ExternalURL().Host,
		},
	})
}

var updateAccountEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: `Update to your Sourcegraph account ({{.Host}})`,
	Text: `
Somebody (likely you) {{.Change}} for the user {{.Username}} on Sourcegraph ({{.Host}}).

If this was not you please change your password immediately.
`,
	HTML: `
<p>
Somebody (likely you) <strong>{{.Change}}</strong> for the user <strong>{{.Username}}</strong> on Sourcegraph ({{.Host}}).
</p>

<p><strong>If this was not you please change your password immediately.</strong></p>
`,
})

// deleteStalePerforceExternalAccounts will remove any Perforce external accounts
// associated with the given user and e-mail combination.
func deleteStalePerforceExternalAccounts(ctx context.Context, db database.DB, userID int32, email string) error {
	if err := db.UserExternalAccounts().Delete(ctx, database.ExternalAccountsDeleteOptions{
		UserID:      userID,
		AccountID:   email,
		ServiceType: extsvc.TypePerforce,
	}); err != nil {
		return errors.Wrap(err, "deleting stale external account")
	}

	// Since we deleted an external account for the user we can no longer trust user
	// based permissions, so we clear them out.
	// This also removes the user's sub-repo permissions.
	if err := db.Authz().RevokeUserPermissions(ctx, &database.RevokeUserPermissionsArgs{UserID: userID}); err != nil {
		return errors.Wrapf(err, "revoking user permissions for user with ID %d", userID)
	}

	return nil

}

// checkEmailAbuse performs abuse prevention checks to prevent email abuse, i.e. users using emails
// of other people whom they want to annoy.
func checkEmailAbuse(ctx context.Context, db database.DB, userID int32) (abused bool, reason string, err error) {
	if conf.EmailVerificationRequired() {
		emails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
			UserID: userID,
		})
		if err != nil {
			return false, "", err
		}

		var verifiedCount, unverifiedCount int
		for _, email := range emails {
			if email.VerifiedAt == nil {
				unverifiedCount++
			} else {
				verifiedCount++
			}
		}

		// Abuse prevention check 1: Require user to have at least one verified email address
		// before adding another.
		//
		// (We need to also allow users who have zero addresses to add one, or else they could
		// delete all emails and then get into an unrecoverable state.)
		//
		// TODO(sqs): prevent users from deleting their last email, when we have the true notion
		// of a "primary" email address.)
		if verifiedCount == 0 && len(emails) != 0 {
			return true, "a verified email is required before you can add additional email addressed to your account", nil
		}

		// Abuse prevention check 2: Forbid user from having many unverified emails to prevent attackers from using this to
		// send spam or a high volume of annoying emails.
		const maxUnverified = 3
		if unverifiedCount >= maxUnverified {
			return true, "too many existing unverified email addresses", nil
		}
	}
	if envvar.SourcegraphDotComMode() {
		// Abuse prevention check 3: Set a quota on Sourcegraph.com users to prevent abuse.
		//
		// There is no quota for on-prem instances because we assume they can trust their users
		// to not abuse adding emails.
		//
		// TODO(sqs): This reuses the "invite quota", which is really just a number that counts
		// down (not specific to invites). Generalize this to just "quota" (remove "invite" from
		// the name).
		if ok, err := db.Users().CheckAndDecrementInviteQuota(ctx, userID); err != nil {
			return false, "", err
		} else if !ok {
			return true, "email address quota exceeded (contact support to increase the quota)", nil
		}
	}
	return false, "", nil
}

// MakeEmailVerificationCode returns a random string that can be used as an email verification
// code. If there is not enough entropy to create a random string, it returns a non-nil error.
func MakeEmailVerificationCode() (string, error) {
	emailCodeBytes := make([]byte, 20)
	if _, err := rand.Read(emailCodeBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(emailCodeBytes), nil
}

// SendUserEmailVerificationEmail sends an email to the user to verify the email address. The code
// is the verification code that the user must provide to verify their access to the email address.
func SendUserEmailVerificationEmail(ctx context.Context, username, email, code string) error {
	q := make(url.Values)
	q.Set("code", code)
	q.Set("email", email)
	verifyEmailPath, _ := router.Router().Get(router.VerifyEmail).URLPath()
	return txemail.Send(ctx, "user_email_verification", txemail.Message{
		To:       []string{email},
		Template: verifyEmailTemplates,
		Data: struct {
			Username string
			URL      string
			Host     string
		}{
			Username: username,
			URL: globals.ExternalURL().ResolveReference(&url.URL{
				Path:     verifyEmailPath.Path,
				RawQuery: q.Encode(),
			}).String(),
			Host: globals.ExternalURL().Host,
		},
	})
}

var verifyEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `Verify your email on Sourcegraph ({{.Host}})`,
	Text: `Hi {{.Username}},

Please verify your email address on Sourcegraph ({{.Host}}) by clicking this link:

{{.URL}}
`,
	HTML: `<p>Hi <a>{{.Username}},</a></p>

<p>Please verify your email address on Sourcegraph ({{.Host}}) by clicking this link:</p>

<p><strong><a href="{{.URL}}">Verify email address</a></p>
`,
})

// triggerPermissionsSync is a helper that attempts to schedule a new permissions
// sync for the given user.
func triggerPermissionsSync(ctx context.Context, logger log.Logger, db database.DB, userID int32, reason string) {
	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		UserIDs: []int32{userID},
		Reason:  reason,
	})
}
