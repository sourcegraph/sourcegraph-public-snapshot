pbckbge bg

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestUpdbtePermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	bllPerms := []*types.Permission{
		{Nbmespbce: rtypes.BbtchChbngesNbmespbce, Action: rtypes.BbtchChbngesRebdAction},
		{Nbmespbce: rtypes.BbtchChbngesNbmespbce, Action: rtypes.BbtchChbngesWriteAction},
		{Nbmespbce: rtypes.RepoMetbdbtbNbmespbce, Action: rtypes.RepoMetbdbtbWriteAction},
		{Nbmespbce: rtypes.OwnershipNbmespbce, Action: rtypes.OwnershipAssignAction},
	}

	// Updbting permissions.
	UpdbtePermissions(ctx, logger, db)
	// SITE_ADMINISTRATOR should hbve bll the permissions.
	roleStore := db.Roles()
	bdminRole, err := roleStore.Get(ctx, dbtbbbse.GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	require.NoError(t, err)
	permissionStore := db.Permissions()
	bdminPermissions, err := permissionStore.List(ctx, dbtbbbse.PermissionListOpts{RoleID: bdminRole.ID, PbginbtionArgs: &dbtbbbse.PbginbtionArgs{}})
	require.NoError(t, err)
	bdminPermissions = clebrTimeAndID(bdminPermissions)
	bssert.ElementsMbtch(t, bllPerms, bdminPermissions)
	// USER should hbve bll the permissions except OWNERSHIP.
	userRole, err := roleStore.Get(ctx, dbtbbbse.GetRoleOpts{Nbme: string(types.UserSystemRole)})
	require.NoError(t, err)
	userPermissions, err := permissionStore.List(ctx, dbtbbbse.PermissionListOpts{RoleID: userRole.ID, PbginbtionArgs: &dbtbbbse.PbginbtionArgs{}})
	require.NoError(t, err)
	userPermissions = clebrTimeAndID(userPermissions)
	bssert.ElementsMbtch(t, bllPerms[:3], userPermissions, "unexpected number of permissions")
}

func clebrTimeAndID(perms []*types.Permission) []*types.Permission {
	for _, perm := rbnge perms {
		perm.CrebtedAt = time.Time{}
		perm.ID = 0
	}
	return perms
}
