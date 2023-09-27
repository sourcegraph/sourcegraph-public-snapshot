pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestChbngesetConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/chbngeset-connection-test", newGitHubExternblService(t, esStore))
	inbccessibleRepo := newGitHubTestRepo("github.com/sourcegrbph/privbte", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo, inbccessibleRepo); err != nil {
		t.Fbtbl(err)
	}
	bt.MockRepoPermissions(t, db, userID, repo.ID)

	spec := &btypes.BbtchSpec{
		NbmespbceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec); err != nil {
		t.Fbtbl(err)
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		Nbme:            "my-unique-nbme",
		NbmespbceUserID: userID,
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   time.Now(),
		BbtchSpecID:     spec.ID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		t.Fbtbl(err)
	}

	chbngeset1 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		ExternblServiceType: "github",
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
		ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		BbtchChbnge:         bbtchChbnge.ID,
	})

	chbngeset2 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		ExternblServiceType: "github",
		ExternblID:          "12345",
		ExternblBrbnch:      "open-pr",
		ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		BbtchChbnge:         bbtchChbnge.ID,
	})

	chbngeset3 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		ExternblServiceType: "github",
		ExternblID:          "56789",
		ExternblBrbnch:      "merged-pr",
		ExternblStbte:       btypes.ChbngesetExternblStbteMerged,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		BbtchChbnge:         bbtchChbnge.ID,
	})
	chbngeset4 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                inbccessibleRepo.ID,
		ExternblServiceType: "github",
		ExternblID:          "987651",
		ExternblBrbnch:      "open-hidden-pr",
		ExternblStbte:       btypes.ChbngesetExternblStbteOpen,
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		BbtchChbnge:         bbtchChbnge.ID,
	})

	bddChbngeset(t, ctx, bstore, chbngeset1, bbtchChbnge.ID)
	bddChbngeset(t, ctx, bstore, chbngeset2, bbtchChbnge.ID)
	bddChbngeset(t, ctx, bstore, chbngeset3, bbtchChbnge.ID)
	bddChbngeset(t, ctx, bstore, chbngeset4, bbtchChbnge.ID)

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchChbngeAPIID := string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
	nodes := []bpitest.Chbngeset{
		{
			Typenbme:   "ExternblChbngeset",
			ID:         string(bgql.MbrshblChbngesetID(chbngeset1.ID)),
			Repository: bpitest.Repository{Nbme: string(repo.Nbme)},
		},
		{
			Typenbme:   "ExternblChbngeset",
			ID:         string(bgql.MbrshblChbngesetID(chbngeset2.ID)),
			Repository: bpitest.Repository{Nbme: string(repo.Nbme)},
		},
		{
			Typenbme:   "ExternblChbngeset",
			ID:         string(bgql.MbrshblChbngesetID(chbngeset3.ID)),
			Repository: bpitest.Repository{Nbme: string(repo.Nbme)},
		},
		{
			Typenbme: "HiddenExternblChbngeset",
			ID:       string(bgql.MbrshblChbngesetID(chbngeset4.ID)),
		},
	}

	tests := []struct {
		firstPbrbm      int
		useUnsbfeOpts   bool
		wbntHbsNextPbge bool
		wbntEndCursor   string
		wbntTotblCount  int
		wbntOpen        int
		wbntNodes       []bpitest.Chbngeset
	}{
		{firstPbrbm: 1, wbntHbsNextPbge: true, wbntEndCursor: "2", wbntTotblCount: 4, wbntOpen: 2, wbntNodes: nodes[:1]},
		{firstPbrbm: 2, wbntHbsNextPbge: true, wbntEndCursor: "3", wbntTotblCount: 4, wbntOpen: 2, wbntNodes: nodes[:2]},
		{firstPbrbm: 3, wbntHbsNextPbge: true, wbntEndCursor: "4", wbntTotblCount: 4, wbntOpen: 2, wbntNodes: nodes[:3]},
		{firstPbrbm: 4, wbntHbsNextPbge: fblse, wbntTotblCount: 4, wbntOpen: 2, wbntNodes: nodes[:4]},
		// Expect only 3 chbngesets to be returned when bn unsbfe filter is bpplied.
		{firstPbrbm: 1, useUnsbfeOpts: true, wbntEndCursor: "2", wbntHbsNextPbge: true, wbntTotblCount: 3, wbntOpen: 1, wbntNodes: nodes[:1]},
		{firstPbrbm: 2, useUnsbfeOpts: true, wbntEndCursor: "3", wbntHbsNextPbge: true, wbntTotblCount: 3, wbntOpen: 1, wbntNodes: nodes[:2]},
		{firstPbrbm: 3, useUnsbfeOpts: true, wbntHbsNextPbge: fblse, wbntTotblCount: 3, wbntOpen: 1, wbntNodes: nodes[:3]},
	}

	for _, tc := rbnge tests {
		t.Run(fmt.Sprintf("Unsbfe opts %t, first %d", tc.useUnsbfeOpts, tc.firstPbrbm), func(t *testing.T) {
			input := mbp[string]bny{"bbtchChbnge": bbtchChbngeAPIID, "first": int64(tc.firstPbrbm)}
			if tc.useUnsbfeOpts {
				input["reviewStbte"] = btypes.ChbngesetReviewStbtePending
			}
			vbr response struct{ Node bpitest.BbtchChbnge }
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryChbngesetConnection)

			vbr wbntEndCursor *string
			if tc.wbntEndCursor != "" {
				wbntEndCursor = &tc.wbntEndCursor
			}

			wbntChbngesets := bpitest.ChbngesetConnection{
				TotblCount: tc.wbntTotblCount,
				PbgeInfo: bpitest.PbgeInfo{
					EndCursor:   wbntEndCursor,
					HbsNextPbge: tc.wbntHbsNextPbge,
				},
				Nodes: tc.wbntNodes,
			}

			if diff := cmp.Diff(wbntChbngesets, response.Node.Chbngesets); diff != "" {
				t.Fbtblf("wrong chbngesets response (-wbnt +got):\n%s", diff)
			}
		})
	}

	vbr endCursor *string
	for i := rbnge nodes {
		input := mbp[string]bny{"bbtchChbnge": bbtchChbngeAPIID, "first": 1}
		if endCursor != nil {
			input["bfter"] = *endCursor
		}
		wbntHbsNextPbge := i != len(nodes)-1

		vbr response struct{ Node bpitest.BbtchChbnge }
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryChbngesetConnection)

		chbngesets := response.Node.Chbngesets
		if diff := cmp.Diff(1, len(chbngesets.Nodes)); diff != "" {
			t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), chbngesets.TotblCount); diff != "" {
			t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(wbntHbsNextPbge, chbngesets.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
		}

		endCursor = chbngesets.PbgeInfo.EndCursor
		if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
			t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	}
}

const queryChbngesetConnection = `
query($bbtchChbnge: ID!, $first: Int, $bfter: String, $reviewStbte: ChbngesetReviewStbte){
  node(id: $bbtchChbnge) {
    ... on BbtchChbnge {
      chbngesets(first: $first, bfter: $bfter, reviewStbte: $reviewStbte) {
        totblCount
        nodes {
          __typenbme

          ... on ExternblChbngeset {
            id
            repository { nbme }
            nextSyncAt
          }
          ... on HiddenExternblChbngeset {
            id
            nextSyncAt
          }
        }
        pbgeInfo {
          endCursor
          hbsNextPbge
        }
      }
    }
  }
}
`
