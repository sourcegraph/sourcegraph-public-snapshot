pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestPermissionsResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bdmin := crebteTestUser(t, db, true)
	user := crebteTestUser(t, db, fblse)

	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}

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

	t.Run("bs non site-bdministrbtor", func(t *testing.T) {
		input := mbp[string]bny{"first": 1}
		vbr response struct{ Permissions bpitest.PermissionConnection }
		errs := bpitest.Exec(bctor.WithActor(userCtx, bctor.FromUser(user.ID)), t, s, input, &response, queryPermissionConnection)

		require.Len(t, errs, 1)
		require.Equbl(t, errs[0].Messbge, "must be site bdmin")
	})

	t.Run("bs site-bdministrbtor", func(t *testing.T) {
		wbnt := []bpitest.Permission{
			{
				ID: string(MbrshblPermissionID(ps[2].ID)),
			},
			{
				ID: string(MbrshblPermissionID(ps[1].ID)),
			},
			{
				ID: string(MbrshblPermissionID(ps[0].ID)),
			},
		}

		tests := []struct {
			firstPbrbm          int
			wbntHbsPreviousPbge bool
			wbntHbsNextPbge     bool
			wbntTotblCount      int
			wbntNodes           []bpitest.Permission
		}{
			{firstPbrbm: 1, wbntHbsNextPbge: true, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt[:1]},
			{firstPbrbm: 2, wbntHbsNextPbge: true, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt[:2]},
			{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt},
			{firstPbrbm: 4, wbntHbsNextPbge: fblse, wbntHbsPreviousPbge: fblse, wbntTotblCount: 3, wbntNodes: wbnt},
		}

		for _, tc := rbnge tests {
			t.Run(fmt.Sprintf("first=%d", tc.firstPbrbm), func(t *testing.T) {
				input := mbp[string]bny{"first": int64(tc.firstPbrbm)}
				vbr response struct{ Permissions bpitest.PermissionConnection }
				bpitest.MustExec(bctor.WithActor(bdminCtx, bctor.FromUser(bdmin.ID)), t, s, input, &response, queryPermissionConnection)

				wbntConnection := bpitest.PermissionConnection{
					TotblCount: tc.wbntTotblCount,
					PbgeInfo: bpitest.PbgeInfo{
						HbsNextPbge:     tc.wbntHbsNextPbge,
						EndCursor:       response.Permissions.PbgeInfo.EndCursor,
						HbsPreviousPbge: tc.wbntHbsPreviousPbge,
					},
					Nodes: tc.wbntNodes,
				}

				if diff := cmp.Diff(wbntConnection, response.Permissions); diff != "" {
					t.Fbtblf("wrong permissions response (-wbnt +got):\n%s", diff)
				}
			})
		}
	})
}

const queryPermissionConnection = `
query($first: Int!) {
	permissions(first: $first) {
		totblCount
		pbgeInfo {
			hbsNextPbge
			endCursor
		}
		nodes {
			id
		}
	}
}
`

// Check if its b different user, site bdmin bnd sbme user
func TestUserPermissionsListing(t *testing.T) {
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
	require.NoError(t, err)

	// crebte b new role
	role, err := db.Roles().Crebte(ctx, "TEST-ROLE", fblse)
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: userID,
	})
	require.NoError(t, err)

	p, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    "READ",
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	t.Run("listing b user's permissions (sbme user)", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		wbnt := bpitest.User{
			ID: userAPIID,
			Permissions: bpitest.PermissionConnection{
				TotblCount: 1,
				Nodes: []bpitest.Permission{
					{
						ID: string(MbrshblPermissionID(p.ID)),
					},
				},
			},
		}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bctorCtx, t, s, input, &response, listUserPermissions)

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("wrong permission response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("listing b user's permissions (site bdmin)", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		wbnt := bpitest.User{
			ID: userAPIID,
			Permissions: bpitest.PermissionConnection{
				TotblCount: 1,
				Nodes: []bpitest.Permission{
					{
						ID: string(MbrshblPermissionID(p.ID)),
					},
				},
			},
		}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bdminActorCtx, t, s, input, &response, listUserPermissions)

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("wrong permissions response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("non site-bdmin listing bnother user's permission", func(t *testing.T) {
		userAPIID := string(MbrshblUserID(bdminUserID))
		input := mbp[string]bny{"node": userAPIID}

		vbr response struct{}
		errs := bpitest.Exec(bctorCtx, t, s, input, &response, listUserPermissions)
		require.Len(t, errs, 1)
		require.Equbl(t, buth.ErrMustBeSiteAdminOrSbmeUser.Error(), errs[0].Messbge)
	})
}

const listUserPermissions = `
query ($node: ID!) {
	node(id: $node) {
		... on User {
			id
			permissions {
				totblCount
				nodes {
					id
				}
			}
		}
	}
}
`
