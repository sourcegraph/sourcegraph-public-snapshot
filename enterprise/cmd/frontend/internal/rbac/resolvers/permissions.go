package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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

func (r *Resolver) Permissions(ctx context.Context, args *gql.ListPermissionArgs) (gql.PermissionConnectionResolver, error) {
	var err error
	var opts = database.PermissionListOpts{}

	if args.Role != nil {
		roleID, err := unmarshalRoleID(*args.Role)
		if err != nil {
			return nil, err
		}

		opts.RoleID = roleID
	}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		opts.UserID = userID
	}

	opts.LimitOffset, err = args.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &permissionConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}
