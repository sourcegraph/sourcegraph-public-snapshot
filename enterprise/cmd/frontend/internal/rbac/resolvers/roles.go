package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *Resolver) Role(ctx context.Context, args *gql.RoleArgs) (gql.RoleResolver, error) {
	return r.roleByID(ctx, args.ID)
}

func (r *Resolver) Roles(ctx context.Context, args *gql.ListRoleArgs) (gql.RoleConnectionResolver, error) {
	var err error
	var opts = database.RolesListOptions{}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: Only viewable for self or by site admins.
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}

		opts.UserID = userID
	}

	opts.LimitOffset, err = args.LimitOffset()
	if err != nil {
		return nil, err
	}

	return &roleConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}

func (r *Resolver) roleByID(ctx context.Context, id graphql.ID) (gql.RoleResolver, error) {
	roleID, err := unmarshalRoleID(id)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, ErrIDIsZero{}
	}

	role, err := r.db.Roles().GetByID(ctx, database.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role}, nil
}
