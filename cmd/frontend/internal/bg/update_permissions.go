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

			permissionsToIncludeForUserRole := collections.NewSet[rtypes.PermissionNamespace](rbacSchema.UserDefaultNamespaces...)
			for _, permission := range permissions {
				rolesToAssign := []types.SystemRole{types.SiteAdministratorSystemRole}
				if permissionsToIncludeForUserRole.Has(permission.Namespace) {
					// After the incident: https://app.incident.io/sourcegraph/incidents/292?source=slack_bookmark,
					// we decided to make the permissions assignment to the user role additive. Thus, developers
					// will have to explicitly state in the rbac schema the permissions they'd like to assign
					// to the user role.
					rolesToAssign = append(rolesToAssign, types.UserSystemRole)
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
