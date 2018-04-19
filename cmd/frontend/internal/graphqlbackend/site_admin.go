package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
)

type createUserResult struct {
	resetPasswordURL string
}

func (r *createUserResult) ResetPasswordURL() string { return r.resetPasswordURL }

func (*schemaResolver) CreateUserBySiteAdmin(ctx context.Context, args *struct {
	Username string
	Email    string
}) (*createUserResult, error) {
	// 🚨 SECURITY: Only site admins can create user accounts.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// The new user will be created with a verified email address.
	user, err := db.Users.Create(ctx, db.NewUser{
		Email:           args.Email,
		EmailIsVerified: true,
		Username:        args.Username,
		Password:        backend.MakeRandomHardToGuessPassword(),
	})
	if err != nil {
		return nil, err
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, user.ID, args.Email)
	if err != nil {
		return nil, err
	}

	return &createUserResult{
		resetPasswordURL: globals.AppURL.ResolveReference(resetURL).String(),
	}, nil
}

type randomizeUserPasswordResult struct {
	resetPasswordURL string
}

func (r *randomizeUserPasswordResult) ResetPasswordURL() string { return r.resetPasswordURL }

func (*schemaResolver) RandomizeUserPasswordBySiteAdmin(ctx context.Context, args *struct {
	User graphql.ID
}) (*randomizeUserPasswordResult, error) {
	// 🚨 SECURITY: Only site admins can randomize user passwords.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	if err := db.Users.RandomizePasswordAndClearPasswordResetRateLimit(ctx, userID); err != nil {
		return nil, err
	}

	email, _, err := db.UserEmails.GetPrimaryEmail(ctx, userID)
	if err != nil {
		return nil, err
	}

	resetURL, err := backend.MakePasswordResetURL(ctx, userID, email)
	if err != nil {
		return nil, err
	}

	return &randomizeUserPasswordResult{
		resetPasswordURL: globals.AppURL.ResolveReference(resetURL).String(),
	}, nil
}

func (*schemaResolver) DeleteUser(ctx context.Context, args *struct {
	User graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser.ID() == args.User {
		return nil, errors.New("unable to delete current user")
	}

	if err := db.Users.Delete(ctx, userID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (*schemaResolver) DeleteOrganization(ctx context.Context, args *struct {
	Organization graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete orgs.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	orgID, err := unmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}

	if err := db.Orgs.Delete(ctx, orgID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (*schemaResolver) SetUserIsSiteAdmin(ctx context.Context, args *struct {
	UserID    graphql.ID
	SiteAdmin bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can promote other users to site admin (or demote from site
	// admin).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.ID() == args.UserID {
		return nil, errors.New("refusing to set current user site admin status")
	}

	userID, err := unmarshalUserID(args.UserID)
	if err != nil {
		return nil, err
	}

	if err := db.Users.SetIsSiteAdmin(ctx, userID, args.SiteAdmin); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
