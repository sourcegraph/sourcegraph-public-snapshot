package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	} else if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil { // 🚨 SECURITY: Only site admins can query all roles.
		return nil, err
	}

	return graphqlutil.NewConnectionResolver[gql.RoleResolver](
		&connectionStore,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			OrderBy: database.OrderBy{
				{Field: "roles.id"},
			},
			Ascending: true,
		},
	)
}

func (r *Resolver) roleByID(ctx context.Context, id graphql.ID) (gql.RoleResolver, error) {
	// 🚨 SECURITY: Only site admins can query role permissions or all permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

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

func (r *Resolver) DeleteRole(ctx context.Context, args *gql.DeleteRoleArgs) (_ *gql.EmptyResponse, err error) {
	// 🚨 SECURITY: Only site administrators can delete roles.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := unmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, ErrIDIsZero{}
	}

	err = r.db.Roles().Delete(ctx, database.DeleteRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) CreateRole(ctx context.Context, args *gql.CreateRoleArgs) (gql.RoleResolver, error) {
	// 🚨 SECURITY: Only site administrators can create roles.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var role *types.Role
	err := r.db.WithTransact(ctx, func(tx database.DB) (err error) {
		role, err = tx.Roles().Create(ctx, args.Name, false)
		if err != nil {
			return err
		}

		if len(args.Permissions) > 0 {
			opts := database.BulkAssignPermissionsToRoleOpts{RoleID: role.ID}
			for _, permissionID := range args.Permissions {
				id, err := unmarshalPermissionID(permissionID)
				if err != nil {
					return err
				}
				opts.Permissions = append(opts.Permissions, id)
			}
			err = tx.RolePermissions().BulkAssignPermissionsToRole(ctx, opts)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &roleResolver{
		db:   r.db,
		role: role,
	}, nil
}

func (r *Resolver) SetRoles(ctx context.Context, args *gql.SetRolesArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site administrators can assign roles to a user.
	// We need to get the current user any
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := gql.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	user, err := auth.CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}

	if user.ID == userID {
		return nil, errors.New("cannot assign role to self")
	}

	opts := database.SetRolesForUserOpts{UserID: userID}

	for _, r := range args.Roles {
		rID, err := unmarshalPermissionID(r)
		if err != nil {
			return nil, err
		}
		opts.Roles = append(opts.Roles, rID)
	}

	if err = r.db.UserRoles().SetRolesForUser(ctx, opts); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}
