package resolvers

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *Resolver) permissionByID(ctx context.Context, id graphql.ID) (gql.PermissionResolver, error) {
	permissionID, err := unmarshalPermissionID(id)
	if err != nil {
		return nil, err
	}

	if permissionID == 0 {
		return nil, ErrIDIsZero{}
	}

	permission, err := r.db.Permissions().GetByID(ctx, database.GetPermissionOpts{
		ID: permissionID,
	})
	if err != nil {
		return nil, err
	}
	return &permissionResolver{permission: permission}, nil
}

func (r *Resolver) Permissions(ctx context.Context, args *gql.ListPermissionArgs) (*graphqlutil.ConnectionResolver[gql.PermissionResolver], error) {
	connectionStore := permisionConnectionStore{
		db: r.db,
	}

	if args.Role != nil {
		// 🚨 SECURITY: Only site admins can query role permissions.
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, err
		}

		roleID, err := unmarshalRoleID(*args.Role)
		if err != nil {
			return nil, err
		}

		if roleID == 0 {
			return nil, errors.New("invalid role id provided")
		}

		connectionStore.roleID = roleID
	}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		if userID == 0 {
			return nil, errors.New("invalid user id provided")
		}

		// 🚨 SECURITY: Only viewable for self or by site admins.
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}

		connectionStore.userID = userID
	}

	return graphqlutil.NewConnectionResolver[gql.PermissionResolver](
		&connectionStore,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			OrderBy: database.OrderBy{
				{Field: "permissions.id"},
			},
		},
	)
}
