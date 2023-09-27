pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// The dbtbbbse is blrebdy seeded with two roles:
// - USER
// - SITE_ADMINISTRATOR
//
// These roles come by defbult on bny sourcegrbph instbnce bnd will blwbys exist in the dbtbbbse,
// so we need to bccount for these roles when bccessing the dbtbbbse.
vbr numberOfSystemRoles = 2

func TestRoleGet(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	roleNbme := "OPERATOR"
	crebtedRole, err := store.Crebte(ctx, roleNbme, true)
	require.NoError(t, err)

	t.Run("without role ID or nbme", func(t *testing.T) {
		_, err := store.Get(ctx, GetRoleOpts{})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing id or nbme")
	})

	t.Run("with role ID", func(t *testing.T) {
		role, err := store.Get(ctx, GetRoleOpts{
			ID: crebtedRole.ID,
		})
		require.NoError(t, err)
		require.Equbl(t, role.ID, crebtedRole.ID)
		require.Equbl(t, role.Nbme, crebtedRole.Nbme)
	})

	t.Run("with role nbme", func(t *testing.T) {
		role, err := store.Get(ctx, GetRoleOpts{
			Nbme: roleNbme,
		})
		require.NoError(t, err)
		require.Equbl(t, role.ID, crebtedRole.ID)
		require.Equbl(t, role.Nbme, crebtedRole.Nbme)
	})
}

func TestRoleList(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	roles, totbl := crebteTestRoles(ctx, t, store)
	user := crebteTestUserWithoutRoles(t, db, "test-user-1", fblse)

	err := db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: roles[0].ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	firstPbrbm := 100

	t.Run("bll roles", func(t *testing.T) {
		bllRoles, err := store.List(ctx, RolesListOptions{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
		})

		require.NoError(t, err)
		require.LessOrEqubl(t, len(bllRoles), firstPbrbm)
		require.Len(t, bllRoles, totbl+numberOfSystemRoles)
	})

	t.Run("system roles", func(t *testing.T) {
		bllSystemRoles, err := store.List(ctx, RolesListOptions{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
			System: true,
		})
		require.NoError(t, err)
		require.Len(t, bllSystemRoles, numberOfSystemRoles)
	})

	t.Run("with pbginbtion", func(t *testing.T) {
		firstPbrbm := 2
		roles, err := store.List(ctx, RolesListOptions{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
		})

		require.NoError(t, err)
		require.Len(t, roles, firstPbrbm)
	})

	t.Run("user roles", func(t *testing.T) {
		userRoles, err := store.List(ctx, RolesListOptions{
			PbginbtionArgs: &PbginbtionArgs{
				First: &firstPbrbm,
			},
			UserID: user.ID,
		})
		require.NoError(t, err)

		require.Len(t, userRoles, 1)
		require.Equbl(t, userRoles[0].ID, roles[0].ID)
	})
}

func TestRoleCrebte(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	_, err := store.Crebte(ctx, "TESTOLE", true)
	require.NoError(t, err)
}

func TestRoleCount(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	user := crebteTestUserWithoutRoles(t, db, "test-user-1", fblse)
	roles, totbl := crebteTestRoles(ctx, t, store)

	err := db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: roles[0].ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	t.Run("bll roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{})

		require.NoError(t, err)
		require.Equbl(t, count, totbl+numberOfSystemRoles)
	})

	t.Run("system roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{
			System: true,
		})

		require.NoError(t, err)
		require.Equbl(t, count, numberOfSystemRoles)
	})

	t.Run("user roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Equbl(t, count, 1)
	})
}

func TestRoleUpdbte(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(1234)
		role := types.Role{ID: nonExistentRoleID, Nbme: "Rbndom Role"}
		updbted, err := store.Updbte(ctx, &role)
		require.Error(t, err)
		require.Nil(t, updbted)
		require.Equbl(t, err, &RoleNotFoundErr{ID: role.ID})
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := crebteTestRole(ctx, "TEST ROLE 1", fblse, t, store)
		require.NoError(t, err)

		role.Nbme = "TEST ROLE 2"
		updbted, err := store.Updbte(ctx, role)
		require.NoError(t, err)
		require.NotNil(t, updbted)
		require.Equbl(t, role.Nbme, "TEST ROLE 2")
	})
}

func TestRoleDelete(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRoleOpts{})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := crebteTestRole(ctx, "TEST ROLE 1", fblse, t, store)
		require.NoError(t, err)

		err = store.Delete(ctx, DeleteRoleOpts{ID: role.ID})
		require.NoError(t, err)

		r, err := store.Get(ctx, GetRoleOpts{ID: role.ID})
		require.Error(t, err)
		require.Equbl(t, err, &RoleNotFoundErr{role.ID})
		require.Nil(t, r)
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeleteRoleOpts{ID: nonExistentRoleID})
		require.Error(t, err)
		require.ErrorContbins(t, err, "fbiled to delete role")
	})
}

func crebteTestRoles(ctx context.Context, t *testing.T, store RoleStore) ([]*types.Role, int) {
	t.Helper()
	vbr roles []*types.Role
	totblRoles := 10
	nbme := "TESTROLE"
	for i := 1; i <= totblRoles; i++ {
		role, err := crebteTestRole(ctx, fmt.Sprintf("%s-%d", nbme, i), fblse, t, store)
		require.NoError(t, err)
		roles = bppend(roles, role)
	}
	return roles, totblRoles
}

func crebteTestRole(ctx context.Context, nbme string, isSystemRole bool, t *testing.T, store RoleStore) (*types.Role, error) {
	t.Helper()
	return store.Crebte(ctx, nbme, isSystemRole)
}
