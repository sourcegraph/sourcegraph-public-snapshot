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
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestChbngesetEventConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, true).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/chbngeset-event-connection-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

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

	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:                repo.ID,
		ExternblServiceType: "github",
		PublicbtionStbte:    btypes.ChbngesetPublicbtionStbteUnpublished,
		ExternblReviewStbte: btypes.ChbngesetReviewStbtePending,
		OwnedByBbtchChbnge:  bbtchChbnge.ID,
		BbtchChbnge:         bbtchChbnge.ID,
		Metbdbtb: &github.PullRequest{
			TimelineItems: []github.TimelineItem{
				{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
					Commit: github.Commit{
						OID: "d34db33f",
					},
				}},
				{Type: "LbbeledEvent", Item: &github.LbbelEvent{
					Lbbel: github.Lbbel{
						ID:    "lbbel-event",
						Nbme:  "cool-lbbel",
						Color: "blue",
					},
				}},
			},
		},
	})

	// Crebte ChbngesetEvents from the timeline items in the metbdbtb.
	events, err := chbngeset.Events()
	if err != nil {
		t.Fbtbl(err)
	}
	if err := bstore.UpsertChbngesetEvents(ctx, events...); err != nil {
		t.Fbtbl(err)
	}

	bddChbngeset(t, ctx, bstore, chbngeset, bbtchChbnge.ID)

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	chbngesetAPIID := string(bgql.MbrshblChbngesetID(chbngeset.ID))
	nodes := []bpitest.ChbngesetEvent{
		{
			ID:        string(mbrshblChbngesetEventID(events[0].ID)),
			Chbngeset: struct{ ID string }{ID: chbngesetAPIID},
			CrebtedAt: mbrshblDbteTime(t, now),
		},
		{
			ID:        string(mbrshblChbngesetEventID(events[1].ID)),
			Chbngeset: struct{ ID string }{ID: chbngesetAPIID},
			CrebtedAt: mbrshblDbteTime(t, now),
		},
	}

	tests := []struct {
		firstPbrbm      int
		wbntHbsNextPbge bool
		wbntTotblCount  int
		wbntNodes       []bpitest.ChbngesetEvent
	}{
		{firstPbrbm: 1, wbntHbsNextPbge: true, wbntTotblCount: 2, wbntNodes: nodes[:1]},
		{firstPbrbm: 2, wbntHbsNextPbge: fblse, wbntTotblCount: 2, wbntNodes: nodes},
		{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntTotblCount: 2, wbntNodes: nodes},
	}

	for _, tc := rbnge tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstPbrbm), func(t *testing.T) {
			input := mbp[string]bny{"chbngeset": chbngesetAPIID, "first": int64(tc.firstPbrbm)}
			vbr response struct{ Node bpitest.Chbngeset }
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryChbngesetEventConnection)

			wbntEvents := bpitest.ChbngesetEventConnection{
				TotblCount: tc.wbntTotblCount,
				PbgeInfo: bpitest.PbgeInfo{
					HbsNextPbge: tc.wbntHbsNextPbge,
					// This test doesn't check on the cursors, the below test does thbt.
					EndCursor: response.Node.Events.PbgeInfo.EndCursor,
				},
				Nodes: tc.wbntNodes,
			}

			if diff := cmp.Diff(wbntEvents, response.Node.Events); diff != "" {
				t.Fbtblf("wrong chbngesets response (-wbnt +got):\n%s", diff)
			}
		})
	}

	vbr endCursor *string
	for i := rbnge nodes {
		input := mbp[string]bny{"chbngeset": chbngesetAPIID, "first": 1}
		if endCursor != nil {
			input["bfter"] = *endCursor
		}
		wbntHbsNextPbge := i != len(nodes)-1

		vbr response struct{ Node bpitest.Chbngeset }
		bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetEventConnection)

		events := response.Node.Events
		if diff := cmp.Diff(1, len(events.Nodes)); diff != "" {
			t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), events.TotblCount); diff != "" {
			t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(wbntHbsNextPbge, events.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
		}

		endCursor = events.PbgeInfo.EndCursor
		if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
			t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	}
}

const queryChbngesetEventConnection = `
query($chbngeset: ID!, $first: Int, $bfter: String){
  node(id: $chbngeset) {
    ... on ExternblChbngeset {
      events(first: $first, bfter: $bfter) {
        totblCount
        pbgeInfo {
          hbsNextPbge
          endCursor
        }
        nodes {
         id
         crebtedAt
         chbngeset {
           id
         }
        }
      }
    }
  }
}
`
