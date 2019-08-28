package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

type UpdateUserPermissionsArgs struct {
	Input *UpdateUserPermissionsInput
}

type UpdateUserPermissionsInput struct {
	Username *string
}

func (*schemaResolver) UpdateUserPermissions(ctx context.Context, args *UpdateUserPermissionsArgs) (*EmptyResponse, error) {
	var username string
	if args.Input.Username != nil {
		username = *args.Input.Username
	}

	user, err := db.Users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the user and site admins are allowed to update the user's permissions.
	err = backend.CheckSiteAdminOrSameUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	_, authzProviders := authz.GetProviders()
	if len(authzProviders) == 0 {
		return &EmptyResponse{}, nil
	}

	for _, p := range authzProviders {
		up, ok := p.(authz.Cache)
		if !ok {
			continue
		}

		err = up.UpdatePermissions(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return &EmptyResponse{}, nil
}
