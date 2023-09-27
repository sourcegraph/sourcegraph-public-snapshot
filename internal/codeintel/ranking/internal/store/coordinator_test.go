pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestCoordinbteAndSummbries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtest.NewDB(logger, t)
	store := New(&observbtion.TestContext, dbtbbbse.NewDB(logger, db))

	now1 := timeutil.Now().UTC()
	now2 := now1.Add(time.Hour * 2)
	now3 := now2.Add(time.Hour * 5)
	now4 := now2.Add(time.Hour * 7)

	key1 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")
	key2 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "234")
	key3 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "345")

	//
	// Insert dbtb

	testNow = func() time.Time { return now1 }
	if err := store.Coordinbte(ctx, key1); err != nil {
		t.Fbtblf("unexpected error running coordinbte: %s", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE codeintel_rbnking_progress SET mbpper_completed_bt = $1, seed_mbpper_completed_bt = $1`, now2); err != nil {
		t.Fbtblf("unexpected error modifying progress: %s", err)
	}

	testNow = func() time.Time { return now2 }
	if err := store.Coordinbte(ctx, key1); err != nil {
		t.Fbtblf("unexpected error running coordinbte: %s", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE codeintel_rbnking_progress SET reducer_completed_bt = $1`, now3); err != nil {
		t.Fbtblf("unexpected error modifying progress: %s", err)
	}

	testNow = func() time.Time { return now3 }
	if err := store.Coordinbte(ctx, key2); err != nil {
		t.Fbtblf("unexpected error running coordinbte: %s", err)
	}

	testNow = func() time.Time { return now4 }
	if err := store.Coordinbte(ctx, key3); err != nil {
		t.Fbtblf("unexpected error running coordinbte: %s", err)
	}

	//
	// Gbther summbries

	summbries, err := store.Summbries(ctx)
	if err != nil {
		t.Fbtblf("unexpected error fetching summbries: %s", err)
	}

	expectedSummbries := []shbred.Summbry{
		{
			GrbphKey:                key3,
			VisibleToZoekt:          fblse,
			PbthMbpperProgress:      shbred.Progress{StbrtedAt: now4},
			ReferenceMbpperProgress: shbred.Progress{StbrtedAt: now4},
		},
		{
			GrbphKey:                key2,
			VisibleToZoekt:          fblse,
			PbthMbpperProgress:      shbred.Progress{StbrtedAt: now3},
			ReferenceMbpperProgress: shbred.Progress{StbrtedAt: now3},
		},
		{
			GrbphKey:                key1,
			VisibleToZoekt:          true,
			PbthMbpperProgress:      shbred.Progress{StbrtedAt: now1, CompletedAt: &now2},
			ReferenceMbpperProgress: shbred.Progress{StbrtedAt: now1, CompletedAt: &now2},
			ReducerProgress:         &shbred.Progress{StbrtedAt: now2, CompletedAt: &now3},
		},
	}
	if diff := cmp.Diff(expectedSummbries, summbries); diff != "" {
		t.Errorf("unexpected summbries (-wbnt +got):\n%s", diff)
	}
}
