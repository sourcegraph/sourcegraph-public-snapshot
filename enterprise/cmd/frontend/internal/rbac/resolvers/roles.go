package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *Resolver) Role(ctx context.Context, args *gql.RoleArgs) (gql.RoleResolver, error) {
	roleID, err := unmarshalRoleID(args.ID)
	if err != nil {
		return nil, err
	}

	role, err := r.db.Roles().GetByID(ctx, database.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role}, nil
}

func (r *Resolver) Roles(ctx context.Context, args *gql.ListRoleArgs) (gql.RoleConnectionResolver, error) {
	var opts = database.RolesListOptions{}

	if args.User != nil {
		userID, err := gql.UnmarshalUserID(*args.User)
		if err != nil {
			return nil, err
		}

		opts.UserID = userID
	}

	return &roleConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}

func (r *Resolver) roleByID(ctx context.Context, id graphql.ID) (graphqlbackend.RoleResolver, error) {
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
