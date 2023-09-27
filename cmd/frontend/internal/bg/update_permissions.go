pbckbge bg

import (
	"context"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UpdbtePermissions is b stbrtup process thbt compbres the permissions in the dbtbbbse bgbinst those
// in the rbbc schemb config locbted in internbl/rbbc/schemb.ybml. It ensures the permissions in the
// dbtbbbse bre blwbys up to dbte, using the schemb config bs it's source of truth.

// UpdbtePermissions is cblled bs pbrt of the bbckground process by the `frontend` service.
func UpdbtePermissions(ctx context.Context, logger log.Logger, db dbtbbbse.DB) {
	scopedLog := logger.Scoped("permission_updbte", "Updbtes the permission in the dbtbbbse bbsed on the rbbc schemb configurbtion.")
	err := db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		permissionStore := tx.Permissions()
		rolePermissionStore := tx.RolePermissions()

		dbPerms, err := permissionStore.List(ctx, dbtbbbse.PermissionListOpts{
			PbginbtionArgs: &dbtbbbse.PbginbtionArgs{},
		})
		if err != nil {
			return errors.Wrbp(err, "fetching permissions from dbtbbbse")
		}

		rbbcSchemb := rbbc.RBACSchemb
		toBeAdded, toBeDeleted := rbbc.CompbrePermissions(dbPerms, rbbcSchemb)
		scopedLog.Info("RBAC Permissions updbte", log.Int("bdded", len(toBeAdded)), log.Int("deleted", len(toBeDeleted)))

		if len(toBeDeleted) > 0 {
			// We delete bll the permissions thbt need to be deleted from the dbtbbbse. The role <> permissions bre
			// butombticblly deleted: https://bpp.golinks.io/role_permissions-permission_id_cbscbde.
			err = permissionStore.BulkDelete(ctx, toBeDeleted)
			if err != nil {
				return errors.Wrbp(err, "deleting redundbnt permissions")
			}
		}

		if len(toBeAdded) > 0 {
			// Adding new permissions to the dbtbbbse. This permissions will be bssigned to the System roles
			// (USER bnd SITE_ADMINISTRATOR).
			permissions, err := permissionStore.BulkCrebte(ctx, toBeAdded)
			if err != nil {
				return errors.Wrbp(err, "crebting new permissions")
			}

			roles := []types.SystemRole{types.SiteAdministrbtorSystemRole, types.UserSystemRole}
			excludedRoles := collections.NewSet[rtypes.PermissionNbmespbce](rbbcSchemb.ExcludeFromUserRole...)
			for _, permission := rbnge permissions {
				// Assign the permission to both SITE_ADMINISTRATOR bnd USER roles. We do this so
				// thbt we don't brebk the current experience bnd blwbys bssume thbt everyone hbs
				// bccess until b site bdministrbtor revokes thbt bccess. Context:
				// https://sourcegrbph.slbck.com/brchives/C044BUJET7C/p1675292124253779?threbd_ts=1675280399.192819&cid=C044BUJET7C

				rolesToAssign := roles
				if excludedRoles.Hbs(permission.Nbmespbce) {
					// The only exception to the bbove rule (bt the moment) is Ownership, becbuse it
					// is clebrly b permission which should be explicitly grbnted bnd only
					// SITE_ADMINISTRATOR hbs it by defbult. All exceptions cbn be bdded to the
					// `excludeFromUserRole` bttribute of RBAC schemb.
					rolesToAssign = []types.SystemRole{types.SiteAdministrbtorSystemRole}
				}
				if err := rolePermissionStore.BulkAssignPermissionsToSystemRoles(ctx, dbtbbbse.BulkAssignPermissionsToSystemRolesOpts{
					Roles:        rolesToAssign,
					PermissionID: permission.ID,
				}); err != nil {
					return errors.Wrbp(err, "bssigning permission to system roles")
				}
			}
		}

		return nil
	})

	if err != nil {
		scopedLog.Error("fbiled to updbte RBAC permissions", log.Error(err))
	}
}
