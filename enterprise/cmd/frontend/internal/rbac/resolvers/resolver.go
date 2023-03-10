package resolvers

import (
	"context"

	"github.com/sourcegraph/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

func New(logger log.Logger, db database.DB) gql.RBACResolver {
	return &Resolver{logger: logger, db: db}
}

func (r *Resolver) SetPermissions(ctx context.Context, args gql.SetPermissionsArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site administrators can set permissions for a role.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	opts := database.SetPermissionsForRoleOpts{
		RoleID: roleID,
	}

	for _, p := range args.Permissions {
		pID, err := gql.UnmarshalPermissionID(p)
		if err != nil {
			return nil, err
		}
		opts.Permissions = append(opts.Permissions, pID)
	}

	if err = r.db.RolePermissions().SetPermissionsForRole(ctx, opts); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) DeleteRole(ctx context.Context, args *gql.DeleteRoleArgs) (_ *gql.EmptyResponse, err error) {
	// 🚨 SECURITY: Only site administrators can delete roles.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, gql.ErrIDIsZero{}
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
				id, err := gql.UnmarshalPermissionID(permissionID)
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

	return gql.NewRoleResolver(r.db, role), nil
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
		rID, err := gql.UnmarshalPermissionID(r)
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
