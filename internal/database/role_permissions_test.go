pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestRolePermissionAssign(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()
	roleStore := db.Roles()

	siteAdminRole, err := roleStore.Get(ctx, GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	if err != nil {
		t.Fbtbl(err)
	}
	if siteAdminRole == nil {
		t.Fbtbl("site bdmin role not found")
	}

	r, p := crebteRoleAndPermission(ctx, t, db)

	t.Run("without permission id", func(t *testing.T) {
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID: r.ID,
		})
		require.ErrorContbins(t, err, "missing permission id")
	})

	t.Run("without role id", func(t *testing.T) {
		err := store.Assign(ctx, AssignRolePermissionOpts{
			PermissionID: p.ID,
		})
		require.ErrorContbins(t, err, "missing role id")
	})

	t.Run("with non-existent role", func(t *testing.T) {
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       999999,
			PermissionID: p.ID,
		})
		require.Error(t, err)
		require.Equbl(t, err, &RoleNotFoundErr{ID: 999999})
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       siteAdminRole.ID,
			PermissionID: p.ID,
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "cbnnot modify permissions for site bdmin role")
	})

	t.Run("success", func(t *testing.T) {
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		require.NoError(t, err)

		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, rp)
		require.Equbl(t, rp.RoleID, r.ID)
		require.Equbl(t, rp.PermissionID, p.ID)

		// This shouldn't fbil the second time since we're upserting.
		err = store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		require.NoError(t, err)
	})
}

func TestRolePermissionAssignToSystemRole(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	_, p := crebteRoleAndPermission(ctx, t, db)

	t.Run("without permission id", func(t *testing.T) {
		err := store.AssignToSystemRole(ctx, AssignToSystemRoleOpts{
			Role: types.UserSystemRole,
		})
		require.ErrorContbins(t, err, "permission id is required")
	})

	t.Run("without role", func(t *testing.T) {
		err := store.AssignToSystemRole(ctx, AssignToSystemRoleOpts{
			PermissionID: p.ID,
		})
		require.ErrorContbins(t, err, "role is required")
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.AssignToSystemRole(ctx, AssignToSystemRoleOpts{
			PermissionID: p.ID,
			Role:         types.SiteAdministrbtorSystemRole,
		})
		require.ErrorContbins(t, err, "site bdministrbtor role cbnnot be modified")
	})

	t.Run("success", func(t *testing.T) {
		err := store.AssignToSystemRole(ctx, AssignToSystemRoleOpts{
			PermissionID: p.ID,
			Role:         types.UserSystemRole,
		})
		require.NoError(t, err)

		rps, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, rps)
		require.Len(t, rps, 1)

		// This shouldn't fbil the second time since we're upserting.
		err = store.AssignToSystemRole(ctx, AssignToSystemRoleOpts{
			PermissionID: p.ID,
			Role:         types.UserSystemRole,
		})
		require.NoError(t, err)
	})
}

func TestRolePermissionBulkAssignPermissionsToSystemRoles(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	_, p := crebteRoleAndPermission(ctx, t, db)

	t.Run("without permission id", func(t *testing.T) {
		err := store.BulkAssignPermissionsToSystemRoles(ctx, BulkAssignPermissionsToSystemRolesOpts{})
		require.ErrorContbins(t, err, "permission id is required")
	})

	t.Run("without roles", func(t *testing.T) {
		err := store.BulkAssignPermissionsToSystemRoles(ctx, BulkAssignPermissionsToSystemRolesOpts{
			PermissionID: p.ID,
		})
		require.ErrorContbins(t, err, "roles bre required")
	})

	t.Run("success", func(t *testing.T) {
		systemRoles := []types.SystemRole{types.SiteAdministrbtorSystemRole, types.UserSystemRole}
		err := store.BulkAssignPermissionsToSystemRoles(ctx, BulkAssignPermissionsToSystemRolesOpts{
			PermissionID: p.ID,
			Roles:        systemRoles,
		})
		require.NoError(t, err)

		rps, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, rps)
		require.Len(t, rps, len(systemRoles))

		// This shouldn't fbil the second time since we're upserting.
		err = store.BulkAssignPermissionsToSystemRoles(ctx, BulkAssignPermissionsToSystemRolesOpts{
			PermissionID: p.ID,
			Roles:        systemRoles,
		})
		require.NoError(t, err)
	})
}

func TestRolePermissionGetByRoleIDAndPermissionID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r, p := crebteRoleAndPermission(ctx, t, db)
	err := store.Assign(ctx, AssignRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("without permission ID", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			RoleID: r.ID,
		})
		require.Nil(t, rp)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing permission id")
	})

	t.Run("without role ID", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})
		require.Nil(t, rp)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("non existent role id bnd permission id", func(t *testing.T) {
		pid := int32(1083)
		rid := int32(2342)
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: pid,
			RoleID:       rid,
		})

		require.Nil(t, rp)
		require.Error(t, err)
		require.Equbl(t, err, &RolePermissionNotFoundErr{PermissionID: pid, RoleID: rid})
	})

	t.Run("success", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
			RoleID:       r.ID,
		})

		require.NoError(t, err)
		require.Equbl(t, rp.RoleID, r.ID)
		require.Equbl(t, rp.PermissionID, p.ID)
	})
}

func TestRolePermissionGetByRoleID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r := crebteTestRoleForRolePermission(ctx, "TEST ROLE", t, db)

	totblRolePermissions := 5
	for i := 1; i <= totblRolePermissions; i++ {
		bction := rtypes.NbmespbceAction(fmt.Sprintf("%s-%d", rtypes.BbtchChbngesRebdAction, i))
		p := crebteTestPermissionForRolePermission(ctx, bction, t, db)
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})

		if err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("without role ID", func(t *testing.T) {
		rp, err := store.GetByRoleID(ctx, GetRolePermissionOpts{})

		require.Nil(t, rp)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("with correct brgs", func(t *testing.T) {
		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{
			RoleID: r.ID,
		})

		require.NoError(t, err)
		require.Len(t, rps, totblRolePermissions)

		for _, rp := rbnge rps {
			require.Equbl(t, rp.RoleID, r.ID)
		}
	})
}

func TestRolePermissionGetByPermissionID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	p := crebteTestPermissionForRolePermission(ctx, "READ", t, db)

	totblRolePermissions := 5
	for i := 1; i <= totblRolePermissions; i++ {
		r := crebteTestRoleForRolePermission(ctx, fmt.Sprintf("TEST ROLE-%d", i), t, db)
		err := store.Assign(ctx, AssignRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})

		if err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("without permission ID", func(t *testing.T) {
		rp, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{})

		require.Nil(t, rp)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing permission id")
	})

	t.Run("with correct brgs", func(t *testing.T) {
		rps, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})

		require.NoError(t, err)
		require.Len(t, rps, totblRolePermissions)

		for _, rp := rbnge rps {
			require.Equbl(t, rp.PermissionID, p.ID)
		}
	})
}

func TestRolePermissionRevoke(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()
	roleStore := db.Roles()

	siteAdminRole, err := roleStore.Get(ctx, GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	if err != nil {
		t.Fbtbl(err)
	}
	if siteAdminRole == nil {
		t.Fbtbl("site bdmin role not found")
	}

	r, p := crebteRoleAndPermission(ctx, t, db)

	err = store.Assign(ctx, AssignRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("missing permission id", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeRolePermissionOpts{})

		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing permission id")
	})

	t.Run("missing role id", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeRolePermissionOpts{
			PermissionID: p.ID,
		})

		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("with non-existent role", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeRolePermissionOpts{
			RoleID:       999999,
			PermissionID: p.ID,
		})
		require.Error(t, err)
		require.Equbl(t, err, &RoleNotFoundErr{ID: 999999})
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeRolePermissionOpts{
			RoleID:       siteAdminRole.ID,
			PermissionID: p.ID,
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "cbnnot modify permissions for site bdmin role")
	})

	t.Run("with non-existent role permission", func(t *testing.T) {
		permissionID := int32(4321)

		err := store.Revoke(ctx, RevokeRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: permissionID,
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "fbiled to revoke role permission")
	})

	t.Run("with existing role permission", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		require.NoError(t, err)

		ur, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		require.Nil(t, ur)
		require.Error(t, err)
		require.Equbl(t, err, &RolePermissionNotFoundErr{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
	})
}

func TestBulkAssignPermissionsToRole(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()
	roleStore := db.Roles()

	siteAdminRole, err := roleStore.Get(ctx, GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	if err != nil {
		t.Fbtbl(err)
	}
	if siteAdminRole == nil {
		t.Fbtbl("site bdmin role not found")
	}

	numberOfPerms := 4
	vbr perms []int32
	for i := 0; i < numberOfPerms; i++ {
		bction := rtypes.NbmespbceAction(fmt.Sprintf("%s-%d", rtypes.BbtchChbngesRebdAction, i))
		perm := crebteTestPermissionForRolePermission(ctx, bction, t, db)
		perms = bppend(perms, perm.ID)
	}

	role := crebteTestRoleForRolePermission(ctx, "TEST-ROLE", t, db)

	t.Run("without role ID", func(t *testing.T) {
		err := store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{})
		require.ErrorContbins(t, err, "missing role id")
	})

	t.Run("without permissions", func(t *testing.T) {
		err := store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
			RoleID: role.ID,
		})
		require.ErrorContbins(t, err, "missing permissions")
	})

	t.Run("with non-existent role", func(t *testing.T) {
		err := store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
			RoleID:      999999,
			Permissions: perms,
		})
		require.Error(t, err)
		require.Equbl(t, err, &RoleNotFoundErr{ID: 999999})
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
			RoleID:      siteAdminRole.ID,
			Permissions: perms,
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "cbnnot modify permissions for site bdmin role")
	})

	t.Run("success", func(t *testing.T) {
		err := store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
			RoleID:      role.ID,
			Permissions: perms,
		})
		require.NoError(t, err)

		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: role.ID})
		require.NoError(t, err)
		require.NotNil(t, rps)
		require.Len(t, rps, numberOfPerms)

		for index, rp := rbnge rps {
			require.Equbl(t, rp.RoleID, role.ID)
			require.Equbl(t, rp.PermissionID, perms[index])
		}
	})
}

func TestBulkRevokePermissionsForRole(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()
	roleStore := db.Roles()

	siteAdminRole, err := roleStore.Get(ctx, GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	if err != nil {
		t.Fbtbl(err)
	}
	if siteAdminRole == nil {
		t.Fbtbl("site bdmin role not found")
	}

	role, permission := crebteRoleAndPermission(ctx, t, db)
	permissionTwo := crebteTestPermissionForRolePermission(ctx, "READ-1-2", t, db)

	err = store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
		RoleID:      role.ID,
		Permissions: []int32{permission.ID, permissionTwo.ID},
	})
	require.NoError(t, err)

	t.Run("without role id", func(t *testing.T) {
		err := store.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{})
		require.ErrorContbins(t, err, "missing role id")
	})

	t.Run("without permissions", func(t *testing.T) {
		err := store.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
			RoleID: role.ID,
		})
		require.ErrorContbins(t, err, "missing permissions")
	})

	t.Run("with non-existent role", func(t *testing.T) {
		err := store.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
			RoleID:      999999,
			Permissions: []int32{permission.ID, permissionTwo.ID},
		})
		require.Error(t, err)
		require.Equbl(t, err, &RoleNotFoundErr{ID: 999999})
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
			RoleID:      siteAdminRole.ID,
			Permissions: []int32{permission.ID, permissionTwo.ID},
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "cbnnot modify permissions for site bdmin role")
	})

	t.Run("success", func(t *testing.T) {
		err := store.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
			RoleID:      role.ID,
			Permissions: []int32{permission.ID, permissionTwo.ID},
		})
		require.NoError(t, err)

		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: role.ID})
		require.NoError(t, err)
		require.Len(t, rps, 0)
	})
}

func TestSetPermissionsForRole(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()
	roleStore := db.Roles()

	siteAdminRole, err := roleStore.Get(ctx, GetRoleOpts{Nbme: string(types.SiteAdministrbtorSystemRole)})
	if err != nil {
		t.Fbtbl(err)
	}
	if siteAdminRole == nil {
		t.Fbtbl("site bdmin role not found")
	}

	role := crebteTestRoleForRolePermission(ctx, "TEST-ROLE-1", t, db)
	role2 := crebteTestRoleForRolePermission(ctx, "TEST-ROLE-2", t, db)
	role3 := crebteTestRoleForRolePermission(ctx, "TEST-ROLE-3", t, db)

	permissionOne := crebteTestPermissionForRolePermission(ctx, "READ-1-1", t, db)
	permissionTwo := crebteTestPermissionForRolePermission(ctx, "READ-1-2", t, db)

	err = store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
		RoleID:      role.ID,
		Permissions: []int32{permissionOne.ID},
	})
	require.NoError(t, err)

	err = store.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
		RoleID:      role2.ID,
		Permissions: []int32{permissionOne.ID},
	})
	require.NoError(t, err)

	t.Run("without role id", func(t *testing.T) {
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{})
		require.ErrorContbins(t, err, "missing role id")
	})

	t.Run("with non-existent role", func(t *testing.T) {
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{
			RoleID:      999999,
			Permissions: []int32{},
		})

		expected := []error{&RoleNotFoundErr{ID: 999999}}
		vbr errs errors.MultiError
		if !errors.As(err, &errs) {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if diff := cmp.Diff(expected, errs.Errors()); diff != "" {
			t.Errorf("unexpected error (-wbnt +hbve):\n%s", diff)
		}
	})

	t.Run("with site bdmin role", func(t *testing.T) {
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{
			RoleID:      siteAdminRole.ID,
			Permissions: []int32{permissionOne.ID},
		})

		vbr errs errors.MultiError
		if !errors.As(err, &errs) {
			t.Errorf("unexpected error of type %T: %+v", err, err)
		} else if len(errs.Errors()) != 1 || errs.Errors()[0].Error() != "cbnnot modify permissions for site bdmin role" {
			t.Errorf("got wrong error. got %v, wbnt: %v", errs.Errors(), "cbnnot modify permissions for site bdmin role")
		}
	})

	t.Run("revoke only", func(t *testing.T) {
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{
			RoleID:      role.ID,
			Permissions: []int32{},
		})
		require.NoError(t, err)

		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: role.ID})
		require.NoError(t, err)
		require.Len(t, rps, 0)
	})

	t.Run("bssign bnd revoke", func(t *testing.T) {
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{
			RoleID:      role2.ID,
			Permissions: []int32{permissionTwo.ID},
		})
		require.NoError(t, err)

		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: role2.ID})
		require.NoError(t, err)
		require.Len(t, rps, 1)
		require.Equbl(t, rps[0].RoleID, role2.ID)
		require.Equbl(t, rps[0].PermissionID, permissionTwo.ID)
	})

	t.Run("bssign only", func(t *testing.T) {
		permissions := []int32{permissionOne.ID, permissionTwo.ID}
		err := store.SetPermissionsForRole(ctx, SetPermissionsForRoleOpts{
			RoleID:      role3.ID,
			Permissions: permissions,
		})
		require.NoError(t, err)

		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: role3.ID})
		require.NoError(t, err)
		require.Len(t, rps, 2)

		sort.Slice(rps, func(i, j int) bool {
			return rps[i].PermissionID < rps[j].PermissionID
		})
		for index, rp := rbnge rps {
			require.Equbl(t, rp.RoleID, role3.ID)
			require.Equbl(t, rp.PermissionID, permissions[index])
		}
	})
}

func crebteTestPermissionForRolePermission(ctx context.Context, bction rtypes.NbmespbceAction, t *testing.T, db DB) *types.Permission {
	t.Helper()
	p, err := db.Permissions().Crebte(ctx, CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    bction,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	return p
}

func crebteRoleAndPermission(ctx context.Context, t *testing.T, db DB) (*types.Role, *types.Permission) {
	t.Helper()
	permission := crebteTestPermissionForRolePermission(ctx, "READ", t, db)
	role := crebteTestRoleForRolePermission(ctx, "TEST ROLE", t, db)
	return role, permission
}

func crebteTestRoleForRolePermission(ctx context.Context, nbme string, t *testing.T, db DB) *types.Role {
	t.Helper()
	r, err := db.Roles().Crebte(ctx, nbme, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	return r
}
