pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRoleConnectionResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := crebteTestUser(t, db, fblse).ID
	userCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	bdminID := crebteTestUser(t, db, true).ID
	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdminID))

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}

	// All sourcegrbph instbnces bre seeded with two system roles bt migrbtion,
	// so we tbke those into bccount when querying roles.
	siteAdminRole, err := db.Roles().Get(ctx, dbtbbbse.GetRoleOpts{
		Nbme: string(types.SiteAdministrbtorSystemRole),
	})
	require.NoError(t, err)

	userRole, err := db.Roles().Get(ctx, dbtbbbse.GetRoleOpts{
		Nbme: string(types.UserSystemRole),
	})
	require.NoError(t, err)

	r, err := db.Roles().Crebte(ctx, "TEST-ROLE", fblse)
	require.NoError(t, err)

	t.Run("bs non site-bdministrbtor", func(t *testing.T) {
		input := mbp[string]bny{"first": 1}
		vbr response struct{ Permissions bpitest.PermissionConnection }
		errs := bpitest.Exec(userCtx, t, s, input, &response, queryRoleConnection)

		require.Len(t, errs, 1)
		require.Equbl(t, errs[0].Messbge, "must be site bdmin")
	})

	t.Run("bs site-bdministrbtor", func(t *testing.T) {
		wbnt := []bpitest.Role{
			{
				ID: string(MbrshblRoleID(userRole.ID)),
			},
			{
				ID: string(MbrshblRoleID(siteAdminRole.ID)),
			},
			{
				ID: string(MbrshblRoleID(r.ID)),
			},
		}

		tests := []struct {
			firstPbrbm          int
			wbntHbsNextPbge     bool
			wbntHbsPreviousPbge bool
			wbntTotblCount      int
			wbntNodes           []bpitest.Role
		}{
			{firstPbrbm: 1, wbntHbsNextPbge: true, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt[:1]},
			{firstPbrbm: 2, wbntHbsNextPbge: true, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt[:2]},
			{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt},
			{firstPbrbm: 4, wbntHbsNextPbge: fblse, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt},
		}

		for _, tc := rbnge tests {
			t.Run(fmt.Sprintf("first=%d", tc.firstPbrbm), func(t *testing.T) {
				input := mbp[string]bny{"first": int64(tc.firstPbrbm)}
				vbr response struct{ Roles bpitest.RoleConnection }
				bpitest.MustExec(bdminCtx, t, s, input, &response, queryRoleConnection)

				wbntConnection := bpitest.RoleConnection{
					TotblCount: tc.wbntTotblCount,
					PbgeInfo: bpitest.PbgeInfo{
						HbsNextPbge:     tc.wbntHbsNextPbge,
						HbsPreviousPbge: tc.wbntHbsPreviousPbge,
					},
					Nodes: tc.wbntNodes,
				}

				if diff := cmp.Diff(wbntConnection, response.Roles); diff != "" {
					t.Fbtblf("wrong roles response (-wbnt +got):\n%s", diff)
				}
			})
		}
	})
}

const queryRoleConnection = `
query($first: Int!) {
	roles(first: $first) {
		totblCount
		pbgeInfo {
			hbsNextPbge
			hbsPreviousPbge
		}
		nodes {
			id
		}
	}
}
`

func TestUserRoleListing(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := crebteTestUser(t, db, fblse).ID
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	bdminUserID := crebteTestUser(t, db, true).ID
	bdminActorCtx := bctor.WithActor(ctx, bctor.FromUser(bdminUserID))

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}
	require.NoError(t, err)

	// crebte b new role
	role, err := db.Roles().Crebte(ctx, "TEST-ROLE", fblse)
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: userID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: bdminUserID,
	})
	require.NoError(t, err)

	t.Run("on sourcegrbph.com", func(t *testing.T) {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		userAPIID := string(MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		vbr response struct{ Node bpitest.User }
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, listUserRoles)
		require.ErrorContbins(t, errs[0], "roles bre not bvbilbble on sourcegrbph.com")
	})

	t.Run("listing b user's roles (sbme user)", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		wbnt := bpitest.User{
			ID: userAPIID,
			Roles: bpitest.RoleConnection{
				TotblCount: 1,
				Nodes: []bpitest.Role{
					{
						ID: string(MbrshblRoleID(role.ID)),
					},
				},
			},
		}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bctorCtx, t, s, input, &response, listUserRoles)

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("wrong role response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("listing b user's roles (site bdmin)", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		wbnt := bpitest.User{
			ID: userAPIID,
			Roles: bpitest.RoleConnection{
				TotblCount: 1,
				Nodes: []bpitest.Role{
					{
						ID: string(MbrshblRoleID(role.ID)),
					},
				},
			},
		}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bdminActorCtx, t, s, input, &response, listUserRoles)

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("wrong roles response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("non site-bdmin listing bnother user's roles", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(bdminUserID))
		input := mbp[string]bny{"node": userAPIID}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bctorCtx, t, s, input, &response, listUserRoles)

		wbnt := bpitest.User{
			ID: userAPIID,
			Roles: bpitest.RoleConnection{
				TotblCount: 1,
				Nodes: []bpitest.Role{
					{
						ID: string(MbrshblRoleID(role.ID)),
					},
				},
			},
		}

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("wrong roles response (-wbnt +got):\n%s", diff)
		}
	})
}

const listUserRoles = `
query ($node: ID!) {
	node(id: $node) {
		... on User {
			id
			roles(first: 50) {
				totblCount
				nodes {
					id
				}
			}
		}
	}
}
`
