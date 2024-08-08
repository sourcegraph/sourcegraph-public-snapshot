package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	iauth "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) CreateUser(ctx context.Context, args *struct {
	Username      string
	Email         *string
	VerifiedEmail *bool
},
) (*createUserResult, error) {
	// 🚨 SECURITY: Only site admins can create user accounts.
	if err := iauth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var email string
	if args.Email != nil {
		email = *args.Email
	}

	// 🚨 SECURITY: Do not assume user email is verified on creation if email delivery is
	// enabled, and we are allowed to reset passwords (which will become the primary
	// mechanism for verifying this newly created email).
	needsEmailVerification := email != "" &&
		conf.CanSendEmail() &&
		userpasswd.ResetPasswordEnabled()
	// For backwards-compatibility, allow this behaviour to be configured based
	// on the VerifiedEmail argument. If not provided, or set to true, we
	// forcibly mark the email as not needing verification.
	if args.VerifiedEmail == nil || *args.VerifiedEmail {
		needsEmailVerification = false
	}

	logger := r.logger.Scoped("createUser").With(
		log.Bool("needsEmailVerification", needsEmailVerification))

	var emailVerificationCode string
	if needsEmailVerification {
		var err error
		emailVerificationCode, err = backend.MakeEmailVerificationCode()
		if err != nil {
			msg := "failed to generate email verification code"
			logger.Error(msg, log.Error(err))
			return nil, errors.Wrap(err, msg)
		}
	}

	user, err := r.db.Users().Create(ctx, database.NewUser{
		Username: args.Username,
		Password: backend.MakeRandomHardToGuessPassword(),

		Email: email,

		// In order to mark an email as unverified, we must generate a verification code.
		EmailIsVerified:       !needsEmailVerification,
		EmailVerificationCode: emailVerificationCode,
	})
	if err != nil {
		msg := "failed to create user"
		logger.Error(msg, log.Error(err))
		return nil, errors.Wrap(err, msg)
	}

	logger = logger.With(log.Int32("userID", user.ID))
	logger.Debug("user created")

	if err = r.db.Authz().GrantPendingPermissions(ctx, &database.GrantPendingPermissionsArgs{
		UserID: user.ID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}); err != nil {
		r.logger.Error("failed to grant user pending permissions",
			log.Error(err))
	}

	return &createUserResult{
		logger:        logger,
		db:            r.db,
		user:          user,
		email:         email,
		emailVerified: !needsEmailVerification,
	}, nil
}

// createUserResult is the result of Mutation.createUser.
//
// 🚨 SECURITY: Only site admins should be able to instantiate this value.
type createUserResult struct {
	logger log.Logger
	db     database.DB

	user          *types.User
	email         string
	emailVerified bool
}

func (r *createUserResult) User(ctx context.Context) *UserResolver {
	return NewUserResolver(ctx, r.db, r.user)
}

// ResetPasswordURL modifies the DB when it generates reset URLs, which is somewhat
// counterintuitive for a "value" type from an implementation POV. Its behavior is
// justified because it is convenient and intuitive from the POV of the API consumer.
func (r *createUserResult) ResetPasswordURL(ctx context.Context) (*string, error) {
	return auth.ResetPasswordURL(ctx, r.db, r.logger, r.user, r.email, r.emailVerified)
}
