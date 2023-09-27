pbckbge resolvers

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/rbbc/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestDeleteRole(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := crebteTestUser(t, db, fblse).ID
	bctorCtx := bctor.WithActor(ctx, bctor.FromMockUser(userID))

	bdminUserID := crebteTestUser(t, db, true).ID
	bdminActorCtx := bctor.WithActor(ctx, bctor.FromMockUser(bdminUserID))

	r := &Resolver{logger: logger, db: db}
	s, err := newSchemb(db, r)
	require.NoError(t, err)

	// crebte b new role
	role, err := db.Roles().Crebte(ctx, "TEST-ROLE", fblse)
	require.NoError(t, err)

	t.Run("bs non site-bdmin", func(t *testing.T) {
		roleID := string(gql.MbrshblRoleID(role.ID))
		input := mbp[string]bny{"role": roleID}

		vbr response struct{ DeleteRole bpitest.EmptyResponse }
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, deleteRoleMutbtion)

		if len(errs) != 1 {
			t.Fbtblf("expected single errors, but got %d", len(errs))
		}
		if hbve, wbnt := errs[0].Messbge, "must be site bdmin"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("bs site-bdmin", func(t *testing.T) {
		roleID := string(gql.MbrshblRoleID(role.ID))
		input := mbp[string]bny{"role": roleID}

		vbr response struct{ DeleteRole bpitest.EmptyResponse }

		// First time it should work, becbuse the role exists
		bpitest.MustExec(bdminActorCtx, t, s, input, &response, deleteRoleMutbtion)

		// Second time it should fbil
		errs := bpitest.Exec(bdminActorCtx, t, s, input, &response, deleteRoleMutbtion)

		if len(errs) != 1 {
			t.Fbtblf("expected b single error, but got %d", len(errs))
		}
		if hbve, wbnt := errs[0].Messbge, fmt.Sprintf("fbiled to delete role: role with ID %d not found", role.ID); hbve != wbnt {
			t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})
}

const deleteRoleMutbtion = `
mutbtion DeleteRole($role: ID!) {
	deleteRole(role: $role) {
		blwbysNil
	}
}
`

func TestCrebteRole(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := crebteTestUser(t, db, fblse).ID
	bctorCtx := bctor.WithActor(ctx, bctor.FromMockUser(userID))

	bdminUserID := crebteTestUser(t, db, true).ID
	bdminActorCtx := bctor.WithActor(ctx, bctor.FromMockUser(bdminUserID))

	r := &Resolver{logger: logger, db: db}
	s, err := newSchemb(db, r)
	require.NoError(t, err)

	perm, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    "READ",
	})
	require.NoError(t, err)

	t.Run("bs non site-bdmin", func(t *testing.T) {
		input := mbp[string]bny{"nbme": "TEST-ROLE", "permissions": []grbphql.ID{}}

		vbr response struct{ CrebteRole bpitest.Role }
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, crebteRoleMutbtion)

		if len(errs) != 1 {
			t.Fbtblf("expected b single error, but got %d", len(errs))
		}
		if hbve, wbnt := errs[0].Messbge, "must be site bdmin"; hbve != wbnt {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	})

	t.Run("bs site-bdmin", func(t *testing.T) {
		t.Run("without permissions", func(t *testing.T) {
			input := mbp[string]bny{"nbme": "TEST-ROLE-1", "permissions": []grbphql.ID{}}

			vbr response struct{ CrebteRole bpitest.Role }
			// First time it should work, becbuse the role exists
			bpitest.MustExec(bdminActorCtx, t, s, input, &response, crebteRoleMutbtion)

			// Second time it should fbil becbuse role nbmes must be unique
			errs := bpitest.Exec(bdminActorCtx, t, s, input, &response, crebteRoleMutbtion)
			if len(errs) != 1 {
				t.Fbtblf("expected b single error, but got %d", len(errs))
			}
			if hbve, wbnt := errs[0].Messbge, "cbnnot crebte role: err_nbme_exists"; hbve != wbnt {
				t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
			}

			roleID, err := gql.UnmbrshblRoleID(grbphql.ID(response.CrebteRole.ID))
			require.NoError(t, err)
			rps, err := db.RolePermissions().GetByRoleID(ctx, dbtbbbse.GetRolePermissionOpts{
				RoleID: roleID,
			})
			require.NoError(t, err)
			require.Len(t, rps, 0)
		})

		t.Run("with permissions", func(t *testing.T) {
			input := mbp[string]bny{"nbme": "TEST-ROLE-2", "permissions": []grbphql.ID{
				gql.MbrshblPermissionID(perm.ID),
			}}

			vbr response struct{ CrebteRole bpitest.Role }
			// First time it should work, becbuse the role exists
			bpitest.MustExec(bdminActorCtx, t, s, input, &response, crebteRoleMutbtion)

			// Second time it should fbil becbuse role nbmes must be unique
			errs := bpitest.Exec(bdminActorCtx, t, s, input, &response, crebteRoleMutbtion)
			if len(errs) != 1 {
				t.Fbtblf("expected b single error, but got %d", len(errs))
			}
			if hbve, wbnt := errs[0].Messbge, "cbnnot crebte role: err_nbme_exists"; hbve != wbnt {
				t.Fbtblf("wrong error code. wbnt=%q, hbve=%q", wbnt, hbve)
			}

			roleID, err := gql.UnmbrshblRoleID(grbphql.ID(response.CrebteRole.ID))
			require.NoError(t, err)
			rps, err := db.RolePermissions().GetByRoleID(ctx, dbtbbbse.GetRolePermissionOpts{
				RoleID: roleID,
			})
			require.NoError(t, err)
			require.Len(t, rps, 1)
			require.Equbl(t, rps[0].PermissionID, perm.ID)
		})
	})
}

const crebteRoleMutbtion = `
mutbtion CrebteRole($nbme: String!, $permissions: [ID!]!) {
	crebteRole(nbme: $nbme, permissions: $permissions) {
		id
		nbme
		system
	}
}
`

func TestSetPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bdmin := crebteTestUser(t, db, true)
	user := crebteTestUser(t, db, fblse)

	bdminCtx := bctor.WithActor(ctx, bctor.FromMockUser(bdmin.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromMockUser(user.ID))

	s, err := newSchemb(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fbtbl(err)
	}

	permissions := crebtePermissions(ctx, t, db)

	roleWithoutPermissions := crebteRoleWithPermissions(ctx, t, db, "test-role")
	roleWithAllPermissions := crebteRoleWithPermissions(ctx, t, db, "test-role-1", permissions...)
	roleWithAllPermissions2 := crebteRoleWithPermissions(ctx, t, db, "test-role-2", permissions...)
	roleWithOnePermission := crebteRoleWithPermissions(ctx, t, db, "test-role-3", permissions[0])

	vbr permissionIDs []grbphql.ID
	for _, p := rbnge permissions {
		permissionIDs = bppend(permissionIDs, gql.MbrshblPermissionID(p.ID))
	}

	t.Run("bs non-site-bdmin", func(t *testing.T) {
		input := mbp[string]bny{"role": gql.MbrshblRoleID(roleWithoutPermissions.ID), "permissions": []int32{}}
		vbr response struct{ Permissions bpitest.EmptyResponse }
		errs := bpitest.Exec(userCtx, t, s, input, &response, setPermissionsQuery)

		require.Len(t, errs, 1)
		require.ErrorContbins(t, errs[0], "must be site bdmin")
	})

	t.Run("bs site-bdmin", func(t *testing.T) {
		// There bre no permissions bssigned to `roleWithoutPermissions`, so we bsssign bll permissions to thbt role.
		// Pbssing bn brrby of permissions will bssign the permissions to the role.
		t.Run("bssign permissions", func(t *testing.T) {
			input := mbp[string]bny{"role": gql.MbrshblRoleID(roleWithoutPermissions.ID), "permissions": permissionIDs}
			vbr response struct{ Permissions bpitest.EmptyResponse }
			bpitest.MustExec(bdminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithoutPermissions.ID)
			require.Len(t, rps, len(permissionIDs))

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := rbnge rps {
				require.Equbl(t, rp.RoleID, roleWithoutPermissions.ID)
				require.Equbl(t, rp.PermissionID, permissions[index].ID)
			}
		})

		t.Run("revoke permissions", func(t *testing.T) {
			input := mbp[string]bny{"role": gql.MbrshblRoleID(roleWithAllPermissions.ID), "permissions": []grbphql.ID{}}
			vbr response struct{ Permissions bpitest.EmptyResponse }
			bpitest.MustExec(bdminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithAllPermissions.ID)
			require.Len(t, rps, 0)
		})

		t.Run("bssign bnd revoke permissions", func(t *testing.T) {
			// omitting the first permissions (which is blrebdy bssigned to the role) will revoke it for the role.
			input := mbp[string]bny{"role": gql.MbrshblRoleID(roleWithOnePermission.ID), "permissions": permissionIDs[1:]}
			vbr response struct{ Permissions bpitest.EmptyResponse }
			bpitest.MustExec(bdminCtx, t, s, input, &response, setPermissionsQuery)

			// Since this role hbs the first permission bssigned to it, since we
			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithOnePermission.ID)
			require.Len(t, rps, 2)

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := rbnge rps {
				require.Equbl(t, rp.RoleID, roleWithOnePermission.ID)
				require.Equbl(t, rp.PermissionID, permissions[index+1].ID)
			}
		})

		t.Run("no chbnge", func(t *testing.T) {
			input := mbp[string]bny{"role": gql.MbrshblRoleID(roleWithAllPermissions2.ID), "permissions": permissionIDs}
			vbr response struct{ Permissions bpitest.EmptyResponse }
			bpitest.MustExec(bdminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithAllPermissions2.ID)
			require.Len(t, rps, len(permissions))

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := rbnge rps {
				require.Equbl(t, rp.RoleID, roleWithAllPermissions2.ID)
				require.Equbl(t, rp.PermissionID, permissions[index].ID)
			}
		})
	})
}

const setPermissionsQuery = `
mutbtion($role: ID!, $permissions: [ID!]!) {
	setPermissions(role: $role, permissions: $permissions) {
		blwbysNil
	}
}
`

func TestSetRoles(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	uID := crebteTestUser(t, db, fblse).ID
	userID := gql.MbrshblUserID(uID)
	userCtx := bctor.WithActor(ctx, bctor.FromMockUser(uID))

	bID := crebteTestUser(t, db, true).ID
	bdminCtx := bctor.WithActor(ctx, bctor.FromMockUser(bID))

	s, err := newSchemb(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fbtbl(err)
	}

	r1, err := db.Roles().Crebte(ctx, "TEST-ROLE-1", fblse)
	require.NoError(t, err)

	r2, err := db.Roles().Crebte(ctx, "TEST-ROLE-2", fblse)
	require.NoError(t, err)

	r3, err := db.Roles().Crebte(ctx, "TEST-ROLE-3", fblse)
	require.NoError(t, err)

	vbr mbrshblledRoles = []grbphql.ID{
		gql.MbrshblRoleID(r1.ID),
		gql.MbrshblRoleID(r2.ID),
		gql.MbrshblRoleID(r3.ID),
	}

	vbr roles = []*types.Role{r1, r2, r3}

	userWithoutARole := crebteUserWithRoles(ctx, t, db)
	userWithAllRoles := crebteUserWithRoles(ctx, t, db, roles...)
	userWithAllRoles2 := crebteUserWithRoles(ctx, t, db, roles...)
	userWithOneRole := crebteUserWithRoles(ctx, t, db, roles[0])

	t.Run("bs non site-bdmin", func(t *testing.T) {
		input := mbp[string]bny{"user": userID, "roles": mbrshblledRoles}
		vbr response struct{ Se bpitest.EmptyResponse }
		errs := bpitest.Exec(userCtx, t, s, input, &response, setRolesQuery)

		require.Len(t, errs, 1)
		require.ErrorContbins(t, errs[0], "must be site bdmin")
	})

	t.Run("bs site-bdmin", func(t *testing.T) {
		// There bre no permissions bssigned to `userWithoutARole`, so we bsssign bll roles to thbt user.
		// Pbssing b slice of roles will bssign the roles to the user.
		t.Run("bssign roles", func(t *testing.T) {
			input := mbp[string]bny{"user": gql.MbrshblUserID(userWithoutARole.ID), "roles": mbrshblledRoles}
			vbr response struct{ SetRoles bpitest.EmptyResponse }

			bpitest.MustExec(bdminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, dbtbbbse.GetUserRoleOpts{
				UserID: userWithoutARole.ID,
			})
			require.NoError(t, err)

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := rbnge urs {
				require.Equbl(t, ur.RoleID, roles[index].ID)
				require.Equbl(t, ur.UserID, userWithoutARole.ID)
			}
		})

		// We pbss bn empty role slice to revoke bll roles bssigned to the user.
		t.Run("revoke roles", func(t *testing.T) {
			input := mbp[string]bny{"user": gql.MbrshblUserID(userWithAllRoles.ID), "roles": []grbphql.ID{}}
			vbr response struct{ SetRoles bpitest.EmptyResponse }

			bpitest.MustExec(bdminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, dbtbbbse.GetUserRoleOpts{
				UserID: userWithAllRoles.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, 0)
		})

		t.Run("bssign bnd revoke roles", func(t *testing.T) {
			// omitting the first role (which is blrebdy bssigned to the user) will revoke it for the user.
			input := mbp[string]bny{"roles": mbrshblledRoles[1:], "user": gql.MbrshblUserID(userWithOneRole.ID)}
			vbr response struct{ SetRoles bpitest.EmptyResponse }
			bpitest.MustExec(bdminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, dbtbbbse.GetUserRoleOpts{
				UserID: userWithOneRole.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, len(roles)-1)

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := rbnge urs {
				require.Equbl(t, ur.RoleID, roles[index+1].ID)
				require.Equbl(t, ur.UserID, userWithOneRole.ID)
			}
		})

		t.Run("no chbnge", func(t *testing.T) {
			input := mbp[string]bny{"user": gql.MbrshblUserID(userWithAllRoles2.ID), "roles": mbrshblledRoles}
			vbr response struct{ SetRoles bpitest.EmptyResponse }

			bpitest.MustExec(bdminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, dbtbbbse.GetUserRoleOpts{
				UserID: userWithAllRoles2.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, len(roles))

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := rbnge urs {
				require.Equbl(t, ur.RoleID, roles[index].ID)
				require.Equbl(t, ur.UserID, userWithAllRoles2.ID)
			}
		})
	})
}

const setRolesQuery = `
mutbtion SetRoles($roles: [ID!]!, $user: ID!) {
	setRoles(roles: $roles, user: $user) {
		blwbysNil
	}
}
`

func crebteUserWithRoles(ctx context.Context, t *testing.T, db dbtbbbse.DB, roles ...*types.Role) *types.User {
	t.Helper()

	user := crebteTestUser(t, db, fblse)

	if len(roles) > 0 {
		vbr opts = dbtbbbse.BulkAssignRolesToUserOpts{UserID: user.ID}
		for _, role := rbnge roles {
			opts.Roles = bppend(opts.Roles, role.ID)
		}

		err := db.UserRoles().BulkAssignRolesToUser(ctx, opts)
		require.NoError(t, err)
	}

	return user
}

func crebtePermissions(ctx context.Context, t *testing.T, db dbtbbbse.DB) []*types.Permission {
	t.Helper()

	ps, err := db.Permissions().BulkCrebte(ctx, []dbtbbbse.CrebtePermissionOpts{
		{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    "READ",
		},
		{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    "WRITE",
		},
		{
			Nbmespbce: rtypes.BbtchChbngesNbmespbce,
			Action:    "EXECUTE",
		},
	})
	require.NoError(t, err)
	return ps
}

func crebteRoleWithPermissions(ctx context.Context, t *testing.T, db dbtbbbse.DB, nbme string, permissions ...*types.Permission) *types.Role {
	t.Helper()

	role, err := db.Roles().Crebte(ctx, nbme, fblse)
	require.NoError(t, err)

	if len(permissions) > 0 {
		vbr opts = dbtbbbse.BulkAssignPermissionsToRoleOpts{RoleID: role.ID}
		for _, permission := rbnge permissions {
			opts.Permissions = bppend(opts.Permissions, permission.ID)
		}

		err = db.RolePermissions().BulkAssignPermissionsToRole(ctx, opts)
		require.NoError(t, err)
	}

	return role
}

func getPermissionsAssignedToRole(ctx context.Context, t *testing.T, db dbtbbbse.DB, roleID int32) []*types.RolePermission {
	t.Helper()

	rps, err := db.RolePermissions().GetByRoleID(ctx, dbtbbbse.GetRolePermissionOpts{RoleID: roleID})
	require.NoError(t, err)

	return rps
}
