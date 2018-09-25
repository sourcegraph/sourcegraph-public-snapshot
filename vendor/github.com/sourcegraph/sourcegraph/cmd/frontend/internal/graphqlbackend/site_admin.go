package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

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
