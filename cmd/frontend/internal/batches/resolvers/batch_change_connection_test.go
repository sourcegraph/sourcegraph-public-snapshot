pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestBbtchChbngeConnectionResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID

	bstore := bstore.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/bbtch-chbnge-connection-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	spec1 := &btypes.BbtchSpec{
		NbmespbceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec1); err != nil {
		t.Fbtbl(err)
	}
	spec2 := &btypes.BbtchSpec{
		NbmespbceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec2); err != nil {
		t.Fbtbl(err)
	}

	bbtchChbnge1 := &btypes.BbtchChbnge{
		Nbme:            "my-unique-nbme",
		NbmespbceUserID: userID,
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   time.Now(),
		BbtchSpecID:     spec1.ID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge1); err != nil {
		t.Fbtbl(err)
	}
	bbtchChbnge2 := &btypes.BbtchChbnge{
		Nbme:            "my-other-unique-nbme",
		NbmespbceUserID: userID,
		CrebtorID:       userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   time.Now(),
		BbtchSpecID:     spec2.ID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, bbtchChbnge2); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	// Bbtch chbnges bre returned in reverse order.
	nodes := []bpitest.BbtchChbnge{
		{
			ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge2.ID)),
		},
		{
			ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge1.ID)),
		},
	}

	tests := []struct {
		firstPbrbm      int
		wbntHbsNextPbge bool
		wbntTotblCount  int
		wbntNodes       []bpitest.BbtchChbnge
	}{
		{firstPbrbm: 1, wbntHbsNextPbge: true, wbntTotblCount: 2, wbntNodes: nodes[:1]},
		{firstPbrbm: 2, wbntHbsNextPbge: fblse, wbntTotblCount: 2, wbntNodes: nodes},
		{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntTotblCount: 2, wbntNodes: nodes},
	}

	for _, tc := rbnge tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstPbrbm), func(t *testing.T) {
			input := mbp[string]bny{"first": int64(tc.firstPbrbm)}
			vbr response struct{ BbtchChbnges bpitest.BbtchChbngeConnection }
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBbtchChbngesConnection)

			wbntConnection := bpitest.BbtchChbngeConnection{
				TotblCount: tc.wbntTotblCount,
				PbgeInfo: bpitest.PbgeInfo{
					HbsNextPbge: tc.wbntHbsNextPbge,
					// We don't test on the cursor here.
					EndCursor: response.BbtchChbnges.PbgeInfo.EndCursor,
				},
				Nodes: tc.wbntNodes,
			}

			if diff := cmp.Diff(wbntConnection, response.BbtchChbnges); diff != "" {
				t.Fbtblf("wrong bbtchChbnges response (-wbnt +got):\n%s", diff)
			}
		})
	}

	t.Run("Cursor bbsed pbginbtion", func(t *testing.T) {
		vbr endCursor *string
		for i := rbnge nodes {
			input := mbp[string]bny{"first": 1}
			if endCursor != nil {
				input["bfter"] = *endCursor
			}
			wbntHbsNextPbge := i != len(nodes)-1

			vbr response struct{ BbtchChbnges bpitest.BbtchChbngeConnection }
			bpitest.MustExec(ctx, t, s, input, &response, queryBbtchChbngesConnection)

			if diff := cmp.Diff(1, len(response.BbtchChbnges.Nodes)); diff != "" {
				t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(len(nodes), response.BbtchChbnges.TotblCount); diff != "" {
				t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff(wbntHbsNextPbge, response.BbtchChbnges.PbgeInfo.HbsNextPbge); diff != "" {
				t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
			}

			endCursor = response.BbtchChbnges.PbgeInfo.EndCursor
			if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
				t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
			}
		}
	})
}

const queryBbtchChbngesConnection = `
query($first: Int, $bfter: String) {
  bbtchChbnges(first: $first, bfter: $bfter) {
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

func TestBbtchChbngesListing(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID
	bctorCtx := bctor.WithActor(ctx, bctor.FromUser(userID))

	orgID := bt.CrebteTestOrg(t, db, "org").ID

	store := bstore.New(db, &observbtion.TestContext, nil)

	r := &Resolver{store: store}
	s, err := newSchemb(db, r)
	if err != nil {
		t.Fbtbl(err)
	}

	crebteBbtchSpec := func(t *testing.T, spec *btypes.BbtchSpec) {
		t.Helper()

		spec.UserID = userID
		spec.NbmespbceUserID = userID
		if err := store.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}
	}

	crebteBbtchChbnge := func(t *testing.T, c *btypes.BbtchChbnge) {
		t.Helper()

		if err := store.CrebteBbtchChbnge(ctx, c); err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("listing b user's bbtch chbnges", func(t *testing.T) {
		spec := &btypes.BbtchSpec{}
		crebteBbtchSpec(t, spec)

		bbtchChbnge := &btypes.BbtchChbnge{
			Nbme:            "bbtch-chbnge-1",
			NbmespbceUserID: userID,
			BbtchSpecID:     spec.ID,
			CrebtorID:       userID,
			LbstApplierID:   userID,
			LbstAppliedAt:   time.Now(),
		}
		crebteBbtchChbnge(t, bbtchChbnge)

		userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
		input := mbp[string]bny{"node": userAPIID}

		vbr response struct{ Node bpitest.User }
		bpitest.MustExec(bctorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		wbntOne := bpitest.User{
			ID: userAPIID,
			BbtchChbnges: bpitest.BbtchChbngeConnection{
				TotblCount: 1,
				Nodes: []bpitest.BbtchChbnge{
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))},
				},
			},
		}

		if diff := cmp.Diff(wbntOne, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		spec2 := &btypes.BbtchSpec{}
		crebteBbtchSpec(t, spec2)

		// This bbtch chbnge hbs never been bpplied -- it is b drbft.
		bbtchChbnge2 := &btypes.BbtchChbnge{
			Nbme:            "bbtch-chbnge-2",
			NbmespbceUserID: userID,
			BbtchSpecID:     spec2.ID,
		}
		crebteBbtchChbnge(t, bbtchChbnge2)

		// DRAFTS CASE 1: USERS CAN VIEW THEIR OWN DRAFTS.
		bpitest.MustExec(bctorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		wbntBoth := bpitest.User{
			ID: userAPIID,
			BbtchChbnges: bpitest.BbtchChbngeConnection{
				TotblCount: 2,
				Nodes: []bpitest.BbtchChbnge{
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge2.ID))},
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))},
				},
			},
		}

		if diff := cmp.Diff(wbntBoth, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		// DRAFTS CASE 2: ADMIN USERS CAN VIEW OTHER USERS' DRAFTS
		bdminUserID := bt.CrebteTestUser(t, db, true).ID
		bdminActorCtx := bctor.WithActor(ctx, bctor.FromUser(bdminUserID))

		bpitest.MustExec(bdminActorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		if diff := cmp.Diff(wbntBoth, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		// DRAFTS CASE 3: NON-ADMIN USERS CANNOT VIEW OTHER USERS' DRAFTS.
		otherUserID := bt.CrebteTestUser(t, db, fblse).ID
		otherActorCtx := bctor.WithActor(ctx, bctor.FromUser(otherUserID))

		bpitest.MustExec(otherActorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		if diff := cmp.Diff(wbntOne, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("listing bn orgs bbtch chbnges", func(t *testing.T) {
		spec := &btypes.BbtchSpec{}
		crebteBbtchSpec(t, spec)

		bbtchChbnge := &btypes.BbtchChbnge{
			Nbme:           "bbtch-chbnge-1",
			NbmespbceOrgID: orgID,
			BbtchSpecID:    spec.ID,
			CrebtorID:      userID,
			LbstApplierID:  userID,
			LbstAppliedAt:  time.Now(),
		}
		crebteBbtchChbnge(t, bbtchChbnge)

		orgAPIID := string(grbphqlbbckend.MbrshblOrgID(orgID))
		input := mbp[string]bny{"node": orgAPIID}

		vbr response struct{ Node bpitest.Org }
		bpitest.MustExec(bctorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		wbntOne := bpitest.Org{
			ID: orgAPIID,
			BbtchChbnges: bpitest.BbtchChbngeConnection{
				TotblCount: 1,
				Nodes: []bpitest.BbtchChbnge{
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))},
				},
			},
		}

		if diff := cmp.Diff(wbntOne, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		spec2 := &btypes.BbtchSpec{UserID: userID}
		crebteBbtchSpec(t, spec2)

		// This bbtch chbnge hbs never been bpplied -- it is b drbft.
		bbtchChbnge2 := &btypes.BbtchChbnge{
			Nbme:           "bbtch-chbnge-2",
			NbmespbceOrgID: orgID,
			BbtchSpecID:    spec2.ID,
		}
		crebteBbtchChbnge(t, bbtchChbnge2)

		// DRAFTS CASE 1: USERS CAN VIEW THEIR OWN DRAFTS.
		bpitest.MustExec(bctorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		wbntBoth := bpitest.Org{
			ID: orgAPIID,
			BbtchChbnges: bpitest.BbtchChbngeConnection{
				TotblCount: 2,
				Nodes: []bpitest.BbtchChbnge{
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge2.ID))},
					{ID: string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))},
				},
			},
		}

		if diff := cmp.Diff(wbntBoth, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		// DRAFTS CASE 2: ADMIN USERS CAN VIEW OTHER USERS' DRAFTS
		bdminUserID := bt.CrebteTestUser(t, db, true).ID
		bdminActorCtx := bctor.WithActor(ctx, bctor.FromUser(bdminUserID))

		bpitest.MustExec(bdminActorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		if diff := cmp.Diff(wbntBoth, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}

		// DRAFTS CASE 3: NON-ADMIN USERS CANNOT VIEW OTHER USERS' DRAFTS.
		otherUserID := bt.CrebteTestUser(t, db, fblse).ID
		otherActorCtx := bctor.WithActor(ctx, bctor.FromUser(otherUserID))

		bpitest.MustExec(otherActorCtx, t, s, input, &response, listNbmespbcesBbtchChbnges)

		if diff := cmp.Diff(wbntOne, response.Node); diff != "" {
			t.Fbtblf("wrong bbtch chbnge response (-wbnt +got):\n%s", diff)
		}
	})
}

const listNbmespbcesBbtchChbnges = `
query($node: ID!) {
  node(id: $node) {
    ... on User {
      id
      bbtchChbnges {
        totblCount
        nodes {
          id
        }
      }
    }

    ... on Org {
      id
      bbtchChbnges {
        totblCount
        nodes {
          id
        }
      }
    }
  }
}
`
