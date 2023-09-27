pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestChbngesetSpecConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)

	bbtchSpec := &btypes.BbtchSpec{
		UserID:          userID,
		NbmespbceUserID: userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	rs := mbke([]*types.Repo, 0, 3)
	for i := 0; i < cbp(rs); i++ {
		nbme := fmt.Sprintf("github.com/sourcegrbph/test-chbngeset-spec-connection-repo-%d", i)
		r := newGitHubTestRepo(nbme, newGitHubExternblService(t, esStore))
		if err := repoStore.Crebte(ctx, r); err != nil {
			t.Fbtbl(err)
		}
		rs = bppend(rs, r)
	}

	chbngesetSpecs := mbke([]*btypes.ChbngesetSpec, 0, len(rs))
	for _, r := rbnge rs {
		repoID := grbphqlbbckend.MbrshblRepositoryID(r.ID)
		s, err := btypes.NewChbngesetSpecFromRbw(bt.NewRbwChbngesetSpecGitBrbnch(repoID, "d34db33f"))
		if err != nil {
			t.Fbtbl(err)
		}
		s.BbtchSpecID = bbtchSpec.ID
		s.UserID = userID
		s.BbseRepoID = r.ID

		if err := bstore.CrebteChbngesetSpec(ctx, s); err != nil {
			t.Fbtbl(err)
		}

		chbngesetSpecs = bppend(chbngesetSpecs, s)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bpiID := string(mbrshblBbtchSpecRbndID(bbtchSpec.RbndID))

	tests := []struct {
		first int

		wbntTotblCount  int
		wbntHbsNextPbge bool
	}{
		{first: 1, wbntTotblCount: 3, wbntHbsNextPbge: true},
		{first: 2, wbntTotblCount: 3, wbntHbsNextPbge: true},
		{first: 3, wbntTotblCount: 3, wbntHbsNextPbge: fblse},
	}

	for _, tc := rbnge tests {
		input := mbp[string]bny{"bbtchSpec": bpiID, "first": tc.first}
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetSpecConnection)

		specs := response.Node.ChbngesetSpecs
		if diff := cmp.Diff(tc.wbntTotblCount, specs.TotblCount); diff != "" {
			t.Fbtblf("first=%d, unexpected totbl count (-wbnt +got):\n%s", tc.first, diff)
		}

		if diff := cmp.Diff(tc.wbntHbsNextPbge, specs.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("first=%d, unexpected hbsNextPbge (-wbnt +got):\n%s", tc.first, diff)
		}
	}

	vbr endCursor *string
	for i := rbnge chbngesetSpecs {
		input := mbp[string]bny{"bbtchSpec": bpiID, "first": 1}
		if endCursor != nil {
			input["bfter"] = *endCursor
		}
		wbntHbsNextPbge := i != len(chbngesetSpecs)-1

		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetSpecConnection)

		specs := response.Node.ChbngesetSpecs
		if diff := cmp.Diff(1, len(specs.Nodes)); diff != "" {
			t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(chbngesetSpecs), specs.TotblCount); diff != "" {
			t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(wbntHbsNextPbge, specs.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
		}

		endCursor = specs.PbgeInfo.EndCursor
		if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
			t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	}
}

const queryChbngesetSpecConnection = `
query($bbtchSpec: ID!, $first: Int!, $bfter: String) {
  node(id: $bbtchSpec) {
    __typenbme

    ... on BbtchSpec {
      id

      chbngesetSpecs(first: $first, bfter: $bfter) {
        totblCount
        pbgeInfo { hbsNextPbge, endCursor }

        nodes {
          __typenbme
          ... on HiddenChbngesetSpec { id }
          ... on VisibleChbngesetSpec { id }
        }
      }
    }
  }
}
`
