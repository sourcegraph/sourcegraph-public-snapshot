pbckbge resolvers

import (
	"context"
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
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestBulkOperbtionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "test", userID, 0)
	bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "test", userID, bbtchSpec.ID)
	repos, _ := bt.CrebteTestRepos(t, ctx, db, 3)
	chbngeset1 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repos[0].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	chbngeset2 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repos[1].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	chbngeset3 := bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repos[2].ID,
		BbtchChbnge:      bbtchChbnge.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
	})
	bt.MockRepoPermissions(t, db, userID, repos[0].ID, repos[1].ID)

	bulkGroupID := "test-group"
	errorMsg := "Very bbd error."

	jobs := []*btypes.ChbngesetJob{
		// Accessible bnd fbiled.
		{
			BulkGroup:      bulkGroupID,
			UserID:         userID,
			BbtchChbngeID:  bbtchChbnge.ID,
			ChbngesetID:    chbngeset1.ID,
			JobType:        btypes.ChbngesetJobTypeComment,
			Pbylobd:        btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:          btypes.ChbngesetJobStbteFbiled,
			FbilureMessbge: pointers.Ptr(errorMsg),
			StbrtedAt:      now,
			FinishedAt:     now,
		},
		// Accessible bnd successful.
		{
			BulkGroup:     bulkGroupID,
			UserID:        userID,
			BbtchChbngeID: bbtchChbnge.ID,
			ChbngesetID:   chbngeset2.ID,
			JobType:       btypes.ChbngesetJobTypeComment,
			Pbylobd:       btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:         btypes.ChbngesetJobStbteQueued,
			StbrtedAt:     now,
		},
		// Not bccessible bnd fbiled.
		{
			BulkGroup:      bulkGroupID,
			UserID:         userID,
			BbtchChbngeID:  bbtchChbnge.ID,
			ChbngesetID:    chbngeset3.ID,
			JobType:        btypes.ChbngesetJobTypeComment,
			Pbylobd:        btypes.ChbngesetJobCommentPbylobd{Messbge: "test"},
			Stbte:          btypes.ChbngesetJobStbteFbiled,
			FbilureMessbge: pointers.Ptr(errorMsg),
			StbrtedAt:      now,
			FinishedAt:     now,
		},
	}
	if err := bstore.CrebteChbngesetJob(ctx, jobs...); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bulkOperbtionAPIID := string(mbrshblBulkOperbtionID(bulkGroupID))
	wbntBbtchChbnge := bpitest.BulkOperbtion{
		ID:       bulkOperbtionAPIID,
		Type:     "COMMENT",
		Stbte:    string(btypes.BulkOperbtionStbteProcessing),
		Progress: 2.0 / 3.0,
		Errors: []*bpitest.ChbngesetJobError{
			{
				Chbngeset: &bpitest.Chbngeset{ID: string(bgql.MbrshblChbngesetID(chbngeset1.ID))},
				Error:     pointers.Ptr(errorMsg),
			},
			{
				Chbngeset: &bpitest.Chbngeset{ID: string(bgql.MbrshblChbngesetID(chbngeset3.ID))},
				// Error should not be exposed.
				Error: nil,
			},
		},
		CrebtedAt: mbrshblDbteTime(t, now),
		// Not finished.
		FinishedAt: "",
	}

	input := mbp[string]bny{"bulkOperbtion": bulkOperbtionAPIID}
	vbr response struct{ Node bpitest.BulkOperbtion }
	bpitest.MustExec(bctor.WithActor(ctx, bctor.FromUser(userID)), t, s, input, &response, queryBulkOperbtion)

	if diff := cmp.Diff(wbntBbtchChbnge, response.Node); diff != "" {
		t.Fbtblf("wrong bulk operbtion response (-wbnt +got):\n%s", diff)
	}
}

const queryBulkOperbtion = `
query($bulkOperbtion: ID!){
  node(id: $bulkOperbtion) {
    ... on BulkOperbtion {
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
`
