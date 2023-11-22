package bg

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/collections"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UpdatePermissions is a startup process that compares the permissions in the database against those
// in the rbac schema config located in internal/rbac/schema.yaml. It ensures the permissions in the
// database are always up to date, using the schema config as it's source of truth.

// UpdatePermissions is called as part of the background process by the `frontend` service.
func UpdatePermissions(ctx context.Context, logger log.Logger, db database.DB) {
	scopedLog := logger.Scoped("permission_update")
	err := db.WithTransact(ctx, func(tx database.DB) error {
		permissionStore := tx.Permissions()
		rolePermissionStore := tx.RolePermissions()

		dbPerms, err := permissionStore.List(ctx, database.PermissionListOpts{
			PaginationArgs: &database.PaginationArgs{},
		})
		if err != nil {
			return errors.Wrap(err, "fetching permissions from database")
		}

		rbacSchema := rbac.RBACSchema
		toBeAdded, toBeDeleted := rbac.ComparePermissions(dbPerms, rbacSchema)
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

			roles := []types.SystemRole{types.SiteAdministratorSystemRole, types.UserSystemRole}
			excludedRoles := collections.NewSet[rtypes.PermissionNamespace](rbacSchema.ExcludeFromUserRole...)
			for _, permission := range permissions {
				// Assign the permission to both SITE_ADMINISTRATOR and USER roles. We do this so
				// that we don't break the current experience and always assume that everyone has
				// access until a site administrator revokes that access. Context:
				// https://sourcegraph.slack.com/archives/C044BUJET7C/p1675292124253779?thread_ts=1675280399.192819&cid=C044BUJET7C

				rolesToAssign := roles
				if excludedRoles.Has(permission.Namespace) {
					// The only exception to the above rule (at the moment) is Ownership, because it
					// is clearly a permission which should be explicitly granted and only
					// SITE_ADMINISTRATOR has it by default. All exceptions can be added to the
					// `excludeFromUserRole` attribute of RBAC schema.
					rolesToAssign = []types.SystemRole{types.SiteAdministratorSystemRole}
				}
				if err := rolePermissionStore.BulkAssignPermissionsToSystemRoles(ctx, database.BulkAssignPermissionsToSystemRolesOpts{
					Roles:        rolesToAssign,
					PermissionID: permission.ID,
				}); err != nil {
					return errors.Wrap(err, "assigning permission to system roles")
				}
			}
		}

		return nil
	})

	if err != nil {
		scopedLog.Error("failed to update RBAC permissions", log.Error(err))
	}
}
