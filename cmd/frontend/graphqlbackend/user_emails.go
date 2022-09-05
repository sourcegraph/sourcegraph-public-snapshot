package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var timeNow = time.Now

func (r *UserResolver) Emails(ctx context.Context) ([]*userEmailResolver, error) {
	// 🚨 SECURITY: Only the authenticated user and site admins can list user's
	// emails.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}

	userEmails, err := r.db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
		UserID: r.user.ID,
	})
	if err != nil {
		return nil, err
	}

	rs := make([]*userEmailResolver, len(userEmails))
	for i, userEmail := range userEmails {
		rs[i] = &userEmailResolver{
			db:        r.db,
			userEmail: *userEmail,
			user:      r,
		}
	}
	return rs, nil
}

type userEmailResolver struct {
	db        database.DB
	userEmail database.UserEmail
	user      *UserResolver
}

func (r *userEmailResolver) Email() string { return r.userEmail.Email }

func (r *userEmailResolver) IsPrimary() bool { return r.userEmail.Primary }

func (r *userEmailResolver) Verified() bool { return r.userEmail.VerifiedAt != nil }
func (r *userEmailResolver) VerificationPending() bool {
	return !r.Verified() && conf.EmailVerificationRequired()
}
func (r *userEmailResolver) User() *UserResolver { return r.user }

func (r *userEmailResolver) ViewerCanManuallyVerify(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: No one can manually verify user's email on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		return false, nil
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == backend.ErrNotAuthenticated || err == backend.ErrMustBeSiteAdmin {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

type addUserEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) AddUserEmail(ctx context.Context, args *addUserEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the authenticated user can add new email to their accounts
	// on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := backend.CheckSameUser(ctx, userID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the authenticated user or site admins can add new email to
		// users' accounts.
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	}

	if err := backend.UserEmails.Add(ctx, r.logger, r.db, userID, args.Email); err != nil {
		return nil, err
	}

	if conf.CanSendEmail() {
		if err := backend.UserEmails.SendUserEmailOnFieldUpdate(ctx, r.logger, r.db, userID, "added an email"); err != nil {
			log15.Warn("Failed to send email to inform user of email addition", "error", err)
		}
	}

	return &EmptyResponse{}, nil
}

type removeUserEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) RemoveUserEmail(ctx context.Context, args *removeUserEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the authenticated user can remove email from their accounts
	// on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := backend.CheckSameUser(ctx, userID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the authenticated user and site admins can remove email
		// from users' accounts.
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	}

	if err := r.db.UserEmails().Remove(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: If an email is removed, invalidate any existing password reset tokens that may have been sent to that email.
	if err := r.db.Users().DeletePasswordResetCode(ctx, userID); err != nil {
		return nil, err
	}

	if conf.CanSendEmail() {
		if err := backend.UserEmails.SendUserEmailOnFieldUpdate(ctx, r.logger, r.db, userID, "removed an email"); err != nil {
			log15.Warn("Failed to send email to inform user of email removal", "error", err)
		}
	}

	return &EmptyResponse{}, nil
}

type setUserEmailPrimaryArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) SetUserEmailPrimary(ctx context.Context, args *setUserEmailPrimaryArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the authenticated user can set the primary email for their
	// accounts on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := backend.CheckSameUser(ctx, userID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the authenticated user and site admins can set the primary
		// email for users' accounts.
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	}

	if err := r.db.UserEmails().SetPrimaryEmail(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	if conf.CanSendEmail() {
		if err := backend.UserEmails.SendUserEmailOnFieldUpdate(ctx, r.logger, r.db, userID, "changed primary email"); err != nil {
			log15.Warn("Failed to send email to inform user of primary address change", "error", err)
		}
	}

	return &EmptyResponse{}, nil
}

type setUserEmailVerifiedArgs struct {
	User     graphql.ID
	Email    string
	Verified bool
}

func (r *schemaResolver) SetUserEmailVerified(ctx context.Context, args *setUserEmailVerifiedArgs) (*EmptyResponse, error) {
	// 🚨 SECURITY: No one can manually verify user's email on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("manually verify user email is disabled")
	}

	// 🚨 SECURITY: Only site admins (NOT users themselves) can manually set email verification
	// status. Users themselves must go through the normal email verification process.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}
	if err := r.db.UserEmails().SetVerified(ctx, userID, args.Email, args.Verified); err != nil {
		return nil, err
	}

	// Avoid unnecessary calls if the email is set to unverified.
	if args.Verified {
		if err = r.db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
			UserID: userID,
			Perm:   authz.Read,
			Type:   authz.PermRepos,
		}); err != nil {
			log15.Error("schemaResolver.SetUserEmailVerified: failed to grant user pending permissions",
				"userID", userID,
				"error", err,
			)
		}
	}

	return &EmptyResponse{}, nil
}

type resendVerificationEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) ResendVerificationEmail(ctx context.Context, args *resendVerificationEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the authenticated user can resend verification email for
	// their accounts on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := backend.CheckSameUser(ctx, userID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the authenticated user and site admins can resend
		// verification email for their accounts.
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	}

	user, err := r.db.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userEmails := r.db.UserEmails()
	lastSent, err := userEmails.GetLatestVerificationSentEmail(ctx, args.Email)
	if err != nil {
		return nil, err
	}
	if lastSent != nil &&
		lastSent.LastVerificationSentAt != nil &&
		timeNow().Sub(*lastSent.LastVerificationSentAt) < 1*time.Minute {
		return nil, errors.New("Last verification email sent too recently")
	}

	email, verified, err := userEmails.Get(ctx, userID, args.Email)
	if err != nil {
		return nil, err
	}
	if verified {
		return &EmptyResponse{}, nil
	}

	code, err := backend.MakeEmailVerificationCode()
	if err != nil {
		return nil, err
	}

	err = userEmails.SetLastVerification(ctx, userID, email, code)
	if err != nil {
		return nil, err
	}

	err = backend.SendUserEmailVerificationEmail(ctx, user.Username, email, code)
	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
