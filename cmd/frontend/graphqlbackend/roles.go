package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) Roles(ctx context.Context, args *ListRoleArgs) (*graphqlutil.ConnectionResolver[RoleResolver], error) {
	connectionStore := roleConnectionStore{
		db:     r.db,
		system: args.System,
	}

	if args.User != nil {
		userID, err := UnmarshalUserID(*args.User)
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
	} else if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil { // 🚨 SECURITY: Only site admins can query all roles.
		return nil, err
	}

	return graphqlutil.NewConnectionResolver[RoleResolver](
		&connectionStore,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			OrderBy: database.OrderBy{
				{Field: "roles.system"},
				{Field: "roles.created_at"},
			},
			Ascending:    false,
			AllowNoLimit: true,
		},
	)
}

func (r *schemaResolver) roleByID(ctx context.Context, id graphql.ID) (RoleResolver, error) {
	// 🚨 SECURITY: Only site admins can query role permissions or all permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := UnmarshalRoleID(id)
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
