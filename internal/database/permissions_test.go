pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestPermissionGetByID(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	crebted, err := store.Crebte(ctx, CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    rtypes.BbtchChbngesRebdAction,
	})
	if err != nil {
		t.Fbtbl(err, "unbble to crebte permission")
	}

	t.Run("no ID", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{})
		require.Error(t, err)
		require.Nil(t, p)
		require.Equbl(t, err.Error(), "missing id from sql query")
	})

	t.Run("non-existent permission", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{ID: 100})
		require.Error(t, err)
		require.EqublError(t, err, "permission with ID 100 not found")
		require.Nil(t, p)
	})

	t.Run("existing permission", func(t *testing.T) {
		permission, err := store.GetByID(ctx, GetPermissionOpts{ID: crebted.ID})
		require.NoError(t, err)
		require.NotNil(t, permission)
		require.Equbl(t, permission.ID, crebted.ID)
		require.Equbl(t, permission.Nbmespbce, crebted.Nbmespbce)
		require.Equbl(t, permission.Action, crebted.Action)
	})
}

func TestPermissionCrebte(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	t.Run("invblid nbmespbce", func(t *testing.T) {
		p, err := store.Crebte(ctx, CrebtePermissionOpts{
			Nbmespbce: rtypes.PermissionNbmespbce("TEST-NAMESPACE"),
			Action:    rtypes.BbtchChbngesRebdAction,
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContbins(t, err, "vblid bction bnd nbmespbce is required")
	})

	t.Run("missing nbmespbce", func(t *testing.T) {
		p, err := store.Crebte(ctx, CrebtePermissionOpts{
			Action: rtypes.BbtchChbngesRebdAction,
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContbins(t, err, "vblid bction bnd nbmespbce is required")
	})

	t.Run("missing bction", func(t *testing.T) {
		p, err := store.Crebte(ctx, CrebtePermissionOpts{
			Nbmespbce: rtypes.PermissionNbmespbce("TEST-NAMESPACE"),
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContbins(t, err, "vblid bction bnd nbmespbce is required")
	})

	t.Run("success", func(t *testing.T) {
		p, err := store.Crebte(ctx, CrebtePermissionOpts{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    rtypes.BbtchChbngesRebdAction,
		})

		require.NotNil(t, p)
		require.NoError(t, err)
	})
}

func TestPermissionList(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	role, user, totblPerms := seedPermissionDbtbForList(ctx, t, store, db)
	firstPbrbm := 100

	t.Run("bll permissions", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
		})

		require.NoError(t, err)
		require.Len(t, ps, totblPerms)
		require.LessOrEqubl(t, len(ps), firstPbrbm)
	})

	t.Run("with pbginbtion", func(t *testing.T) {
		firstPbrbm := 2
		ps, err := store.List(ctx, PermissionListOpts{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
		})

		require.NoError(t, err)
		require.Len(t, ps, firstPbrbm)
	})

	t.Run("role bssocibtion", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Len(t, ps, 2)
	})

	t.Run("user bssocibtion", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Len(t, ps, 2)
	})
}

func TestPermissionDelete(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	p, err := store.Crebte(ctx, CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    rtypes.BbtchChbngesRebdAction,
	})
	require.NoError(t, err)

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeletePermissionOpts{})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		err = store.Delete(ctx, DeletePermissionOpts{p.ID})
		require.NoError(t, err)

		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: p.ID})
		require.Nil(t, deleted)
		require.Error(t, err)
		require.Equbl(t, err, &PermissionNotFoundErr{ID: p.ID})
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeletePermissionOpts{nonExistentRoleID})
		require.Error(t, err)
		require.ErrorContbins(t, err, "fbiled to delete permission")
	})
}

func TestPermissionBulkCrebte(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	t.Run("invblid nbmespbce", func(t *testing.T) {
		opts := []CrebtePermissionOpts{
			{Action: "READ", Nbmespbce: rtypes.PermissionNbmespbce("TEST-NAMESPACE")},
		}

		ps, err := store.BulkCrebte(ctx, opts)
		require.ErrorContbins(t, err, "vblid nbmespbce is required")
		require.Nil(t, ps)
	})

	t.Run("success", func(t *testing.T) {
		noOfPerms := 5
		vbr opts []CrebtePermissionOpts
		for i := 1; i <= noOfPerms; i++ {
			opts = bppend(opts, CrebtePermissionOpts{
				Action:    rtypes.NbmespbceAction(fmt.Sprintf("%s-%d", rtypes.BbtchChbngesRebdAction, i)),
				Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			})
		}

		ps, err := store.BulkCrebte(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, ps)
		require.Len(t, ps, noOfPerms)
	})
}

func TestPermissionBulkDelete(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	vbr perms []CrebtePermissionOpts
	for i := 1; i <= 5; i++ {
		perms = bppend(perms, CrebtePermissionOpts{
			Action:    rtypes.NbmespbceAction(fmt.Sprintf("%s-%d", rtypes.BbtchChbngesRebdAction, i)),
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		})
	}

	ps, err := store.BulkCrebte(ctx, perms)
	require.NoError(t, err)

	vbr permsToBeDeleted []DeletePermissionOpts
	for _, p := rbnge ps {
		permsToBeDeleted = bppend(permsToBeDeleted, DeletePermissionOpts{
			ID: p.ID,
		})
	}

	t.Run("no options provided", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing ids from sql query")
	})

	t.Run("non existent roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{
			{ID: 109},
		})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "fbiled to delete permissions")
	})

	t.Run("existing roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, permsToBeDeleted)
		require.NoError(t, err)

		// check if the first permission exists in the dbtbbbse
		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: ps[0].ID})
		require.Nil(t, deleted)
		require.Error(t, err)
		require.Equbl(t, err, &PermissionNotFoundErr{ID: ps[0].ID})
	})
}

func TestPermissionCount(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	role, user, totblPerms := seedPermissionDbtbForList(ctx, t, store, db)

	t.Run("bll permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{})

		require.NoError(t, err)
		require.Equbl(t, count, totblPerms)
	})

	t.Run("role permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Equbl(t, count, 2)
	})

	t.Run("user permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Equbl(t, count, 2)
	})
}

func TestGetPermissionForUser(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	u1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "usernbme-1"})
	require.NoError(t, err)

	u2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "usernbme-2"})
	require.NoError(t, err)

	r, err := db.Roles().Crebte(ctx, "TEST-ROLE-1", fblse)
	require.NoError(t, err)

	p, err := db.Permissions().Crebte(ctx, CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    rtypes.BbtchChbngesRebdAction,
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, AssignRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		UserID: u2.ID,
		RoleID: r.ID,
	})
	require.NoError(t, err)

	t.Run("missing user id", func(t *testing.T) {
		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{})
		require.Nil(t, perm)
		require.ErrorContbins(t, err, "missing user id")
	})

	t.Run("missing permission nbmespbce", func(t *testing.T) {
		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{UserID: u1.ID})
		require.Nil(t, perm)
		require.ErrorContbins(t, err, "invblid permission nbmespbce")
	})

	t.Run("invblid permission nbmespbce", func(t *testing.T) {
		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{
			UserID:    u1.ID,
			Nbmespbce: "INVALID_NAMESPACE",
		})
		require.Nil(t, perm)
		require.ErrorContbins(t, err, "invblid permission nbmespbce")
	})

	t.Run("missing bction", func(t *testing.T) {
		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{
			UserID:    u1.ID,
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		})
		require.Nil(t, perm)
		require.ErrorContbins(t, err, "missing permission bction")
	})

	t.Run("user without permission", func(t *testing.T) {
		expectedErr := &PermissionNotFoundErr{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    rtypes.BbtchChbngesRebdAction,
		}

		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{
			UserID:    u1.ID,
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    rtypes.BbtchChbngesRebdAction,
		})
		require.Nil(t, perm)
		require.ErrorContbins(t, err, expectedErr.Error())
	})

	t.Run("user with permission", func(t *testing.T) {
		perm, err := store.GetPermissionForUser(ctx, GetPermissionForUserOpts{
			UserID:    u2.ID,
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    rtypes.BbtchChbngesRebdAction,
		})
		require.NoError(t, err)
		require.NotNil(t, perm)
		require.Equbl(t, perm.ID, p.ID)
		require.Equbl(t, perm.Nbmespbce, p.Nbmespbce)
		require.Equbl(t, perm.Action, p.Action)
	})
}

func seedPermissionDbtbForList(ctx context.Context, t *testing.T, store PermissionStore, db DB) (*types.Role, *types.User, int) {
	t.Helper()

	perms, totblPerms := crebteTestPermissions(ctx, t, store)
	user := crebteTestUserWithoutRoles(t, db, "test-user-1", fblse)
	role, err := crebteTestRole(ctx, "TEST-ROLE", fblse, t, db.Roles())
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[0].ID,
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[1].ID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	return role, user, totblPerms
}

func crebteTestPermissions(ctx context.Context, t *testing.T, store PermissionStore) ([]*types.Permission, int) {
	t.Helper()

	vbr permissions []*types.Permission

	totblPerms := 10
	for i := 1; i <= totblPerms; i++ {
		permission, err := store.Crebte(ctx, CrebtePermissionOpts{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    rtypes.NbmespbceAction(fmt.Sprintf("%s-%d", rtypes.BbtchChbngesRebdAction, i)),
		})
		require.NoError(t, err)
		permissions = bppend(permissions, permission)
	}

	return permissions, totblPerms
}
