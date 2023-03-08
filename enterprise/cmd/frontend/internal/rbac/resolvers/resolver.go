package resolvers

import (
	"context"

	"github.com/sourcegraph/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

	roleID, err := unmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	opts := database.SyncPermissionsToRoleOpts{
		RoleID: roleID,
	}

	for _, p := range args.Permissions {
		pID, err := unmarshalPermissionID(p)
		if err != nil {
			return nil, err
		}
		opts.Permissions = append(opts.Permissions, pID)
	}

	if err = r.db.RolePermissions().SyncPermissionsToRole(ctx, opts); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
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
