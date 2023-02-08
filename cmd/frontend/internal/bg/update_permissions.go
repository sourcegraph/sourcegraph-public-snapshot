package bg

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UpdatePermissions is a startup process that compares the permissions in the database against those
// in the rbac schema config located in internal/rbac/schema.yaml. It ensures the permissions in the
// database are always up to date, using the schema config as it's source of truth.

// This method is called as part of the background process by the `frontend` service.
func UpdatePermissions(ctx context.Context, logger log.Logger, db database.DB) {
	scopedLog := logger.Scoped("permission_update", "Updates the permission in the database based on the rbac schema configuration.")
	err := db.WithTransact(ctx, func(tx database.DB) error {
		permissionStore := tx.Permissions()
		roleStore := tx.Roles()
		rolePermissionStore := tx.RolePermissions()

		dbPerms, err := permissionStore.FetchAll(ctx)
		if err != nil {
			return errors.Wrap(err, "fetching permissions from database")
		}

		toBeAdded, toBeDeleted := rbac.ComparePermissions(dbPerms, rbac.RBACSchema)
		scopedLog.Info("RBAC Permissions update", log.Int("added", len(toBeAdded)), log.Int("deleted", len(toBeDeleted)))

		if len(toBeDeleted) > 0 {
			// We delete all the permissions that need to be deleted from the database. The role <> permissions are
			// automatically deleted: https://app.golinks.io/role_permissions-permission_id_cascade.
			err = permissionStore.BulkDelete(ctx, toBeDeleted)
			if err != nil {
				return errors.Wrap(err, "deleting redundant permissions")
			}
		}

		if len(toBeAdded) > 0 {
			// Adding new permissions to the database. This permissions will be assigned to the System roles
			// (USER and SITE_ADMINISTRATOR).
			permissions, err := permissionStore.BulkCreate(ctx, toBeAdded)
			if err != nil {
				return errors.Wrap(err, "creating new permissions")
			}

			// Currently, we have only two system roles so we can just list the first two. In the future,
			// it might be worth creating a new method called `FetchAll` or `ListWithoutPagination` to
			// retrieve all system roles, but since we know currently there won't be more than two system
			// roles at any given point in time, then this works.
			firstParam := 2
			systemRoles, err := roleStore.List(ctx, database.RolesListOptions{
				PaginationArgs: &database.PaginationArgs{
					First: &firstParam,
				},
				System: true,
			})
			if err != nil {
				return errors.Wrap(err, "fetching system roles")
			}

			for _, permission := range permissions {
				for _, role := range systemRoles {
					_, err := rolePermissionStore.Create(ctx, database.CreateRolePermissionOpts{
						PermissionID: permission.ID,
						RoleID:       role.ID,
					})
					if err != nil {
						return errors.Wrapf(err, "assigning permission to role: %s", role.Name)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		scopedLog.Error("failed to update RBAC permissions", log.Error(err))
	}
}
