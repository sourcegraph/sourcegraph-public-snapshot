package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) CreateUser(ctx context.Context, args *struct {
	Username string
	Email    *string
}) (*createUserResult, error) {
	// 🚨 SECURITY: Only site admins can create user accounts.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var email string
	if args.Email != nil {
		email = *args.Email
	}

	// 🚨 SECURITY: Do not assume user email is verified on creation if email delivery is
	// enabled and we are allowed to reset passwords (which will become the primary
	// mechanism for verifying this newly created email).
	var verificationCode string
	if email != "" {
		var err error
		verificationCode, err = backend.MakeEmailVerificationCode()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate email verification code")
		}
	}

	user, err := r.db.Users().Create(ctx, database.NewUser{
		Username: args.Username,
		Password: backend.MakeRandomHardToGuessPassword(),

		Email: email,

		// In order to mark an email as unverified, we must generate a verification code.
		EmailIsVerified:       verificationCode == "",
		EmailVerificationCode: verificationCode,
	})
	if err != nil {
		return nil, err
	}

	logger := r.logger.Scoped("createUser", "create user handler").With(
		log.Bool("requireVerification", email != ""),
		log.Int32("user.id", user.ID),
	)
	logger.Debug("user created")

	if err = r.db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
		UserID: user.ID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}); err != nil {
		r.logger.Error("failed to grant user pending permissions",
			log.Error(err))
	}

	return &createUserResult{db: r.db, user: user, emailWasProvided: email != ""}, nil
}

// createUserResult is the result of Mutation.createUser.
//
// 🚨 SECURITY: Only site admins should be able to instantiate this value.
type createUserResult struct {
	db database.DB

	user             *types.User
	emailWasProvided bool
}

func (r *createUserResult) User() *UserResolver { return NewUserResolver(r.db, r.user) }

func (r *createUserResult) ResetPasswordURL(ctx context.Context) (*string, error) {
	if !userpasswd.ResetPasswordEnabled() {
		return nil, nil
	}

	var ru string
	if conf.CanSendEmail() && r.emailWasProvided {
		ru, err := userpasswd.HandleSetPasswordEmail(ctx, r.db, r.user.ID)
		if err != nil {
			return nil, err
		}
		return &ru, nil
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, r.db, r.user.ID)
	if err != nil {
		return nil, err
	}
	ru = globals.ExternalURL().ResolveReference(resetURL).String()
	return &ru, nil
}
