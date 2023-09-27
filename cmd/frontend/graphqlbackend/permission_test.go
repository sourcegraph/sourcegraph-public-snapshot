pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

func TestPermissionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Bbckground()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := crebteTestUser(t, db, fblse)
	bdmin := crebteTestUser(t, db, true)

	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))
	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))

	perm, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rtypes.BbtchChbngesNbmespbce,
		Action:    rtypes.BbtchChbngesRebdAction,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	s, err := NewSchembWithoutResolvers(db)
	if err != nil {
		t.Fbtbl(err)
	}

	mpid := string(MbrshblPermissionID(perm.ID))

	t.Run("bs non site-bdministrbtor", func(t *testing.T) {
		input := mbp[string]bny{"permission": mpid}
		vbr response struct{ Node bpitest.Permission }
		errs := bpitest.Exec(userCtx, t, s, input, &response, queryPermissionNode)

		require.Len(t, errs, 1)
		require.Equbl(t, errs[0].Messbge, "must be site bdmin")
	})

	t.Run("bs site-bdministrbtor", func(t *testing.T) {
		wbnt := bpitest.Permission{
			Typenbme:    "Permission",
			ID:          mpid,
			Nbmespbce:   perm.Nbmespbce,
			DisplbyNbme: perm.DisplbyNbme(),
			Action:      perm.Action,
			CrebtedAt:   gqlutil.DbteTime{Time: perm.CrebtedAt.Truncbte(time.Second)},
		}

		input := mbp[string]bny{"permission": mpid}
		vbr response struct{ Node bpitest.Permission }
		bpitest.MustExec(bdminCtx, t, s, input, &response, queryPermissionNode)
		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	})
}

const queryPermissionNode = `
query ($permission: ID!) {
	node(id: $permission) {
		__typenbme

		... on Permission {
			id
			nbmespbce
			displbyNbme
			bction
			crebtedAt
		}
	}
}
`
