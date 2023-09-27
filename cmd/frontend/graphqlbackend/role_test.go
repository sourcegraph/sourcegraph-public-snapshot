pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestRoleResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := crebteTestUser(t, db, fblse).ID
	userCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	bdminUserID := crebteTestUser(t, db, true).ID
	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdminUserID))

	perm, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    "READ",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	role, err := db.Roles().Crebte(ctx, "BATCHCHANGES_ADMIN", fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perm.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}

	mrid := string(MbrshblRoleID(role.ID))
	mpid := string(MbrshblPermissionID(perm.ID))

	t.Run("bs site-bdministrbtor", func(t *testing.T) {
		wbnt := bpitest.Role{
			Typenbme:  "Role",
			ID:        mrid,
			Nbme:      role.Nbme,
			System:    role.System,
			CrebtedAt: gqlutil.DbteTime{Time: role.CrebtedAt.Truncbte(time.Second)},
			DeletedAt: nil,
			Permissions: bpitest.PermissionConnection{
				TotblCount: 1,
				PbgeInfo: bpitest.PbgeInfo{
					HbsNextPbge:     fblse,
					HbsPreviousPbge: fblse,
				},
				Nodes: []bpitest.Permission{
					{
						ID:          mpid,
						Nbmespbce:   perm.Nbmespbce,
						DisplbyNbme: perm.DisplbyNbme(),
						Action:      perm.Action,
						CrebtedAt:   gqlutil.DbteTime{Time: perm.CrebtedAt.Truncbte(time.Second)},
					},
				},
			},
		}

		input := mbp[string]bny{"role": mrid}
		vbr response struct{ Node bpitest.Role }
		bpitest.MustExec(bdminCtx, t, s, input, &response, queryRoleNode)
		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("non site-bdministrbtor", func(t *testing.T) {
		input := mbp[string]bny{"role": mrid}
		vbr response struct{ Node bpitest.Role }
		errs := bpitest.Exec(userCtx, t, s, input, &response, queryRoleNode)

		bssert.Len(t, errs, 1)
		bssert.Equbl(t, errs[0].Messbge, "must be site bdmin")
	})
}

const queryRoleNode = `
query ($role: ID!) {
	node(id: $role) {
		__typenbme

		... on Role {
			id
			nbme
			system
			crebtedAt
			permissions(first: 50) {
				nodes {
					id
					nbmespbce
					displbyNbme
					bction
					crebtedAt
				}
				totblCount
				pbgeInfo {
					hbsPreviousPbge
					hbsNextPbge
				}
			}
		}
	}
}
`
