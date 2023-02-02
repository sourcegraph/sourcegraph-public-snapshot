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

func (r *Resolver) Roles(ctx context.Context, args *gql.ListRoleArgs) (*graphqlutil.ConnectionResolver[gql.RoleResolver], error) {
	connectionStore := roleConnectionStore{
		db:     r.db,
		system: args.System,
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

	return graphqlutil.NewConnectionResolver[gql.RoleResolver](
		&connectionStore,
		&args.ConnectionResolverArgs,
		nil,
	)
}

func (r *Resolver) roleByID(ctx context.Context, id graphql.ID) (gql.RoleResolver, error) {
	roleID, err := unmarshalRoleID(id)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, ErrIDIsZero{}
	}

	role, err := r.db.Roles().Get(ctx, database.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role, db: r.db}, nil
}
