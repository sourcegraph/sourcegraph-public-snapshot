package graphqlbackend

import (
	"context"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

func (u *userResolver) Emails(ctx context.Context) ([]*userEmailResolver, error) {
	// 🚨 SECURITY: Only the self user and site admins can fetch a user's emails.
	if err := backend.CheckSiteAdminOrSameUser(ctx, u.user.ID); err != nil {
		return nil, err
	}

	userEmails, err := db.UserEmails.ListByUser(ctx, u.user.ID)
	if err != nil {
		return nil, err
	}

	rs := make([]*userEmailResolver, len(userEmails))
	for i, userEmail := range userEmails {
		rs[i] = &userEmailResolver{
			userEmail: *userEmail,
			user:      u,
		}
	}
	return rs, nil
}

type userEmailResolver struct {
	userEmail db.UserEmail
	user      *userResolver
}

func (r *userEmailResolver) Email() string  { return r.userEmail.Email }
func (r *userEmailResolver) Verified() bool { return r.userEmail.VerifiedAt != nil }
func (r *userEmailResolver) VerificationPending() bool {
	return !r.Verified() && conf.EmailVerificationRequired()
}
func (r *userEmailResolver) User() *userResolver { return r.user }

func (r *userEmailResolver) ViewerCanManuallyVerify(ctx context.Context) (bool, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err == backend.ErrNotAuthenticated || err == backend.ErrMustBeSiteAdmin {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *schemaResolver) AddUserEmail(ctx context.Context, args *struct {
	User  graphql.ID
	Email string
}) (*EmptyResponse, error) {
	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins can add an email address to a user.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		return nil, err
	}

	// TODO(sqs): this should send an email verification email to the user, but that code is in the
	// ../app package and needs to be extracted.
	if err := db.UserEmails.Add(ctx, userID, args.Email, nil); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) SetUserEmailVerified(ctx context.Context, args *struct {
	User     graphql.ID
	Email    string
	Verified bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins (NOT users themselves) can manually set email verification
	// status. Users themselves must go through the normal email verification process.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}
	if err := db.UserEmails.SetVerified(ctx, userID, args.Email, args.Verified); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
