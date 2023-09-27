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
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestBulkOperbtionConnectionResolver(t *testing.T) {
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

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test", userID, bbtchSpec.ID)
	bbtchChbngeAPIID := bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID)
	repos, _ := bt.CrebteTestRepos(t, ctx, db, 3)
	chbngeset := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repos[0].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	jobs := []*btypes.ChbngesetJob{
		{
			BulkGroup:     "group-1",
			UserID:        userID,
			BbtchChbngeID: bbtchChbnge.ID,
			ChbngesetID:   chbngeset.ID,
			JobType:       btypes.ChbngesetJobTypeComment,
			Pbylobd:       btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:         btypes.ChbngesetJobStbteQueued,
			StbrtedAt:     now,
			FinishedAt:    now,
		},
		{
			BulkGroup:     "group-2",
			UserID:        userID,
			BbtchChbngeID: bbtchChbnge.ID,
			ChbngesetID:   chbngeset.ID,
			JobType:       btypes.ChbngesetJobTypeComment,
			Pbylobd:       btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:         btypes.ChbngesetJobStbteQueued,
			StbrtedAt:     now,
			FinishedAt:    now,
		},
		{
			BulkGroup:     "group-3",
			UserID:        userID,
			BbtchChbngeID: bbtchChbnge.ID,
			ChbngesetID:   chbngeset.ID,
			JobType:       btypes.ChbngesetJobTypeComment,
			Pbylobd:       btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:         btypes.ChbngesetJobStbteQueued,
			StbrtedAt:     now,
			FinishedAt:    now,
		},
	}
	if err := bstore.CrebteChbngesetJob(ctx, jobs...); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	nodes := []bpitest.BulkOperbtion{
		{
			ID:        string(mbrshblBulkOperbtionID("group-3")),
			Type:      "COMMENT",
			Stbte:     string(btypes.BulkOperbtionStbteProcessing),
			Errors:    []*bpitest.ChbngesetJobError{},
			CrebtedAt: mbrshblDbteTime(t, now),
		},
		{
			ID:        string(mbrshblBulkOperbtionID("group-2")),
			Type:      "COMMENT",
			Stbte:     string(btypes.BulkOperbtionStbteProcessing),
			Errors:    []*bpitest.ChbngesetJobError{},
			CrebtedAt: mbrshblDbteTime(t, now),
		},
		{
			ID:        string(mbrshblBulkOperbtionID("group-1")),
			Type:      "COMMENT",
			Stbte:     string(btypes.BulkOperbtionStbteProcessing),
			Errors:    []*bpitest.ChbngesetJobError{},
			CrebtedAt: mbrshblDbteTime(t, now),
		},
	}

	tests := []struct {
		firstPbrbm      int
		wbntHbsNextPbge bool
		wbntEndCursor   string
		wbntTotblCount  int
		wbntNodes       []bpitest.BulkOperbtion
	}{
		{firstPbrbm: 1, wbntHbsNextPbge: true, wbntEndCursor: "2", wbntTotblCount: 3, wbntNodes: nodes[:1]},
		{firstPbrbm: 2, wbntHbsNextPbge: true, wbntEndCursor: "1", wbntTotblCount: 3, wbntNodes: nodes[:2]},
		{firstPbrbm: 3, wbntHbsNextPbge: fblse, wbntTotblCount: 3, wbntNodes: nodes[:3]},
	}

	for _, tc := rbnge tests {
		t.Run(fmt.Sprintf("First %d", tc.firstPbrbm), func(t *testing.T) {
			input := mbp[string]bny{"bbtchChbnge": bbtchChbngeAPIID, "first": int64(tc.firstPbrbm)}
			vbr response struct {
				Node bpitest.BbtchChbnge
			}
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBulkOperbtionConnection)

			vbr wbntEndCursor *string
			if tc.wbntEndCursor != "" {
				wbntEndCursor = &tc.wbntEndCursor
			}

			wbntBulkOperbtions := bpitest.BulkOperbtionConnection{
				TotblCount: tc.wbntTotblCount,
				PbgeInfo: bpitest.PbgeInfo{
					EndCursor:   wbntEndCursor,
					HbsNextPbge: tc.wbntHbsNextPbge,
				},
				Nodes: tc.wbntNodes,
			}

			if diff := cmp.Diff(wbntBulkOperbtions, response.Node.BulkOperbtions); diff != "" {
				t.Fbtblf("wrong bulk operbtions response (-wbnt +got):\n%s", diff)
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

		vbr response struct {
			Node bpitest.BbtchChbnge
		}
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBulkOperbtionConnection)

		bulkOperbtions := response.Node.BulkOperbtions
		if diff := cmp.Diff(1, len(bulkOperbtions.Nodes)); diff != "" {
			t.Fbtblf("unexpected number of nodes (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(len(nodes), bulkOperbtions.TotblCount); diff != "" {
			t.Fbtblf("unexpected totbl count (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff(wbntHbsNextPbge, bulkOperbtions.PbgeInfo.HbsNextPbge); diff != "" {
			t.Fbtblf("unexpected hbsNextPbge (-wbnt +got):\n%s", diff)
		}

		endCursor = bulkOperbtions.PbgeInfo.EndCursor
		if wbnt, hbve := wbntHbsNextPbge, endCursor != nil; hbve != wbnt {
			t.Fbtblf("unexpected endCursor existence. wbnt=%t, hbve=%t", wbnt, hbve)
		}
	}
}

const queryBulkOperbtionConnection = `
query($bbtchChbnge: ID!, $first: Int, $bfter: String){
    node(id: $bbtchChbnge) {
        ... on BbtchChbnge {
            bulkOperbtions(first: $first, bfter: $bfter) {
                totblCount
                pbgeInfo {
                    endCursor
                    hbsNextPbge
                }
                nodes {
                    id
                    type
                    stbte
                    progress
                    errors {
                        chbngeset {
                            id
                        }
                        error
                    }
                    crebtedAt
                    finishedAt
                }
            }
        }
    }
}
`
