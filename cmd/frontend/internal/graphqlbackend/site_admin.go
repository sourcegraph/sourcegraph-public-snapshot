package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (*schemaResolver) SetUserIsSiteAdmin(ctx context.Context, args *struct {
	UserID    graphql.ID
	SiteAdmin bool
}) (*EmptyResponse, error) {
	if !actor.FromContext(ctx).IsAdmin() {
		return nil, errors.New("must be site admin to set users as site admins")
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

	if err := localstore.Users.SetIsSiteAdmin(ctx, userID, args.SiteAdmin); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
