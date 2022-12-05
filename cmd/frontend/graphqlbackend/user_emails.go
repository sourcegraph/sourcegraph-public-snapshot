package graphqlbackend

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var timeNow = time.Now

func (r *UserResolver) Emails(ctx context.Context) ([]*userEmailResolver, error) {
	// 🚨 SECURITY: Only the authenticated user and site admins can list user's
	// emails.
	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == auth.ErrNotAuthenticated || err == auth.ErrMustBeSiteAdmin {
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

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.Add(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	if conf.CanSendEmail() {
		if err := userEmails.SendUserEmailOnFieldUpdate(ctx, userID, "added an email"); err != nil {
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

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.Remove(ctx, userID, args.Email); err != nil {
		return nil, err
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

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.SetPrimaryEmail(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type setUserEmailVerifiedArgs struct {
	User     graphql.ID
	Email    string
	Verified bool
}

func (r *schemaResolver) SetUserEmailVerified(ctx context.Context, args *setUserEmailVerifiedArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.SetVerified(ctx, userID, args.Email, args.Verified); err != nil {
		return nil, err
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

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.ResendVerificationEmail(ctx, userID, args.Email, timeNow()); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
