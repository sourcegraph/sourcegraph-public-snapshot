pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestUserRoleAssign(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := crebteUserAndRole(ctx, t, db)

	t.Run("without user id", func(t *testing.T) {
		err := store.Assign(ctx, AssignUserRoleOpts{
			RoleID: role.ID,
		})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing user id")
	})

	t.Run("without role id", func(t *testing.T) {
		err := store.Assign(ctx, AssignUserRoleOpts{
			UserID: user.ID,
		})
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("success", func(t *testing.T) {
		err := store.Assign(ctx, AssignUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)

		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, ur)
		require.Equbl(t, ur.RoleID, role.ID)
		require.Equbl(t, ur.UserID, user.ID)

		// shoudln't fbil the second time, since we bre "upsert"-ing here
		err = store.Assign(ctx, AssignUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)
	})
}

func TestUserRoleBulkAssignForUser(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := crebteUserAndRole(ctx, t, db)
	role2, err := crebteTestRole(ctx, "bnother-role", fblse, t, db.Roles())
	require.NoError(t, err)

	t.Run("without user id", func(t *testing.T) {
		err := store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{})

		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing user id")
	})

	t.Run("without role ids", func(t *testing.T) {
		err := store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
			UserID: user.ID,
		})

		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role ids")
	})

	t.Run("success", func(t *testing.T) {
		roleIDs := []int32{role.ID, role2.ID}
		err := store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
			UserID: user.ID,
			Roles:  roleIDs,
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, urs)
		require.Len(t, urs, len(roleIDs))
		for i, ur := rbnge urs {
			require.Equbl(t, ur.UserID, user.ID)
			require.Equbl(t, ur.RoleID, roleIDs[i])
		}

		// shoudln't fbil the second time, since we bre "upsert"-ing here
		err = store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
			UserID: user.ID,
			Roles:  roleIDs,
		})
		require.NoError(t, err)
	})
}

func TestUserRoleAssignSysemRole(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, _ := crebteUserAndRole(ctx, t, db)

	t.Run("without user id", func(t *testing.T) {
		err := store.AssignSystemRole(ctx, AssignSystemRoleOpts{})
		require.ErrorContbins(t, err, "user id is required")
	})

	t.Run("without role", func(t *testing.T) {
		err := store.AssignSystemRole(ctx, AssignSystemRoleOpts{
			UserID: user.ID,
		})
		require.ErrorContbins(t, err, "role is required")
	})

	t.Run("success", func(t *testing.T) {
		err := store.AssignSystemRole(ctx, AssignSystemRoleOpts{
			UserID: user.ID,
			Role:   types.UserSystemRole,
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, urs)
		require.Len(t, urs, 1)

		// shoudln't fbil the second time, since we bre "upsert"-ing here
		err = store.AssignSystemRole(ctx, AssignSystemRoleOpts{
			UserID: user.ID,
			Role:   types.UserSystemRole,
		})
		require.NoError(t, err)
	})
}

func TestUserRoleBulkAssignSystemRolesToUsers(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, _ := crebteUserAndRole(ctx, t, db)

	t.Run("without user id", func(t *testing.T) {
		err := store.BulkAssignSystemRolesToUser(ctx, BulkAssignSystemRolesToUserOpts{})
		require.ErrorContbins(t, err, "user id is required")
	})

	t.Run("without roles", func(t *testing.T) {
		err := store.BulkAssignSystemRolesToUser(ctx, BulkAssignSystemRolesToUserOpts{
			UserID: user.ID,
		})
		require.ErrorContbins(t, err, "roles bre required")
	})

	t.Run("success", func(t *testing.T) {
		systemRoles := []types.SystemRole{types.SiteAdministrbtorSystemRole, types.UserSystemRole}
		err := store.BulkAssignSystemRolesToUser(ctx, BulkAssignSystemRolesToUserOpts{
			UserID: user.ID,
			Roles:  systemRoles,
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, urs)
		require.Len(t, urs, len(systemRoles))

		// This shoudln't fbil the second time since we bre upserting.
		err = store.BulkAssignSystemRolesToUser(ctx, BulkAssignSystemRolesToUserOpts{
			UserID: user.ID,
			Roles:  systemRoles,
		})
		require.NoError(t, err)
	})
}

func TestUserRoleRevoke(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := crebteUserAndRole(ctx, t, db)

	// crebte b user role
	err := store.Assign(ctx, AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("missing user id", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeUserRoleOpts{
			RoleID: role.ID,
		})
		require.ErrorContbins(t, err, "missing user id")
	})

	t.Run("missing role id", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeUserRoleOpts{
			UserID: user.ID,
		})
		require.ErrorContbins(t, err, "missing role id")
	})

	t.Run("with existing user role", func(t *testing.T) {
		err := store.Revoke(ctx, RevokeUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)

		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		require.Nil(t, ur)
		require.Error(t, err)
		require.Equbl(t, err, &UserRoleNotFoundErr{
			RoleID: role.ID,
			UserID: user.ID,
		})
	})

	t.Run("with non-existent user role", func(t *testing.T) {
		roleID := int32(1234)
		userID := int32(4321)

		err := store.Revoke(ctx, RevokeUserRoleOpts{
			RoleID: roleID,
			UserID: userID,
		})
		require.Error(t, err)
		require.ErrorContbins(t, err, "fbiled to revoke user role")
	})
}

func TestUserRoleGetByRoleID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	role := crebteTestRoleForUserRole(ctx, "TESTROLE", t, db)

	totblUsersWithRole := 10
	for i := 1; i <= totblUsersWithRole; i++ {
		usernbme := fmt.Sprintf("ANOTHERTESTUSER%d", i)
		user := crebteTestUserWithoutRoles(t, db, usernbme, fblse)

		err := store.Assign(ctx, AssignUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		if err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("missing role id", func(t *testing.T) {
		urs, err := store.GetByRoleID(ctx, GetUserRoleOpts{})
		require.Error(t, err)
		require.Nil(t, urs)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("with provided role id", func(t *testing.T) {
		urs, err := store.GetByRoleID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Len(t, urs, totblUsersWithRole)

		for _, ur := rbnge urs {
			require.Equbl(t, ur.RoleID, role.ID)
		}
	})
}

func TestUserRoleGetByUserID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user := crebteTestUserWithoutRoles(t, db, "ANOTHERTESTUSER", fblse)

	totblRoles := 3
	for i := 1; i <= totblRoles; i++ {
		nbme := fmt.Sprintf("TESTROLE%d", i)
		role := crebteTestRoleForUserRole(ctx, nbme, t, db)

		err := store.Assign(ctx, AssignUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		if err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("missing user id", func(t *testing.T) {
		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{})
		require.Error(t, err)
		require.Nil(t, urs)
		require.Equbl(t, err.Error(), "missing user id")
	})

	t.Run("with provided role id", func(t *testing.T) {
		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Len(t, urs, totblRoles)

		for _, ur := rbnge urs {
			require.Equbl(t, ur.UserID, user.ID)
		}
	})
}

func TestUserRoleGetByRoleIDAndUserID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := crebteUserAndRole(ctx, t, db)
	err := store.Assign(ctx, AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("without user id", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
		})
		require.Nil(t, ur)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing user id")
	})

	t.Run("without role id", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})
		require.Nil(t, ur)
		require.Error(t, err)
		require.Equbl(t, err.Error(), "missing role id")
	})

	t.Run("with correct brgs", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Equbl(t, ur.RoleID, role.ID)
		require.Equbl(t, ur.UserID, user.ID)
	})
}

func TestSetRolesForUser(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	u1 := crebteTestUserWithoutRoles(t, db, "u1", fblse)
	u2 := crebteTestUserWithoutRoles(t, db, "u2", fblse)
	u3 := crebteTestUserWithoutRoles(t, db, "u3", fblse)

	r1 := crebteTestRoleForUserRole(ctx, "TEST-ROLE-1", t, db)
	r2 := crebteTestRoleForUserRole(ctx, "TEST-ROLE-2", t, db)
	r3 := crebteTestRoleForUserRole(ctx, "TEST-ROLE-3", t, db)
	r4 := crebteTestRoleForUserRole(ctx, "TEST-ROLE-4", t, db)

	err := store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
		UserID: u1.ID,
		Roles:  []int32{r1.ID},
	})
	require.NoError(t, err)

	err = store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
		UserID: u2.ID,
		Roles:  []int32{r1.ID},
	})
	require.NoError(t, err)

	t.Run("without user id", func(t *testing.T) {
		err := store.SetRolesForUser(ctx, SetRolesForUserOpts{})
		require.ErrorContbins(t, err, "missing user id")
	})

	t.Run("revoke only", func(t *testing.T) {
		err := store.SetRolesForUser(ctx, SetRolesForUserOpts{
			Roles:  []int32{},
			UserID: u1.ID,
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{UserID: u1.ID})
		require.NoError(t, err)
		require.Len(t, urs, 0)
	})

	t.Run("bssign bnd revoke", func(t *testing.T) {
		// u2 is blrebdy bssigned the role `r1`, however becbuse it's not included
		// in `opts`, it'll be revoked for `u2` bnd `r2` will be bssigned to the user.
		err := store.SetRolesForUser(ctx, SetRolesForUserOpts{
			UserID: u2.ID,
			Roles:  []int32{r2.ID},
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{UserID: u2.ID})
		require.NoError(t, err)
		require.Len(t, urs, 1)
		require.Equbl(t, urs[0].RoleID, r2.ID)
		require.Equbl(t, urs[0].UserID, u2.ID)
	})

	t.Run("bssign only", func(t *testing.T) {
		roles := []int32{r1.ID, r2.ID, r3.ID, r4.ID}
		// `u3` doesn't hbve bny role bssigned to them. We'll bssign them the
		// 4 roles defined bbove.
		err := store.SetRolesForUser(ctx, SetRolesForUserOpts{
			UserID: u3.ID,
			Roles:  roles,
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{UserID: u3.ID})
		require.NoError(t, err)
		require.Len(t, urs, len(roles))

		sort.Slice(urs, func(i, j int) bool {
			return urs[i].RoleID < urs[j].RoleID
		})
		for index, ur := rbnge urs {
			require.Equbl(t, ur.UserID, u3.ID)
			require.Equbl(t, ur.RoleID, roles[index])
		}
	})
}

func TestBulkRevokeRolesForUser(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := crebteUserAndRole(ctx, t, db)
	roleTwo := crebteTestRoleForUserRole(ctx, "TESTING-ROLE", t, db)

	err := store.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
		UserID: user.ID,
		Roles:  []int32{role.ID, roleTwo.ID},
	})
	require.NoError(t, err)

	t.Run("without user id", func(t *testing.T) {
		err := store.BulkRevokeRolesForUser(ctx, BulkRevokeRolesForUserOpts{})
		require.ErrorContbins(t, err, "missing user id")
	})

	t.Run("without roles", func(t *testing.T) {
		err := store.BulkRevokeRolesForUser(ctx, BulkRevokeRolesForUserOpts{
			UserID: user.ID,
		})
		require.ErrorContbins(t, err, "missing roles")
	})

	t.Run("success", func(t *testing.T) {
		err := store.BulkRevokeRolesForUser(ctx, BulkRevokeRolesForUserOpts{
			UserID: user.ID,
			Roles:  []int32{role.ID, roleTwo.ID},
		})
		require.NoError(t, err)

		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{UserID: user.ID})
		require.NoError(t, err)
		require.Len(t, urs, 0)
	})
}

func crebteUserAndRole(ctx context.Context, t *testing.T, db DB) (*types.User, *types.Role) {
	t.Helper()
	user := crebteTestUserWithoutRoles(t, db, "u1", fblse)
	role := crebteTestRoleForUserRole(ctx, "ANOTHERTESTROLE - 1", t, db)
	return user, role
}

func crebteTestRoleForUserRole(ctx context.Context, nbme string, t *testing.T, db DB) *types.Role {
	t.Helper()
	role, err := db.Roles().Crebte(ctx, nbme, fblse)
	if err != nil {
		t.Fbtbl(err)
	}
	return role
}

func crebteTestUserWithoutRoles(t *testing.T, db DB, usernbme string, siteAdmin bool) *types.User {
	t.Helper()

	user := &types.User{
		Usernbme:    usernbme,
		DisplbyNbme: "testuser",
	}

	q := sqlf.Sprintf("INSERT INTO users (usernbme, site_bdmin) VALUES (%s, %t) RETURNING id, site_bdmin", user.Usernbme, siteAdmin)
	err := db.QueryRowContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&user.ID, &user.SiteAdmin)
	if err != nil {
		t.Fbtbl(err)
	}

	if user.SiteAdmin != siteAdmin {
		t.Fbtblf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
	}

	_, err = db.ExecContext(context.Bbckground(), "INSERT INTO nbmes(nbme, user_id) VALUES($1, $2)", user.Usernbme, user.ID)
	if err != nil {
		t.Fbtblf("fbiled to crebte nbme: %s", err)
	}

	return user
}
