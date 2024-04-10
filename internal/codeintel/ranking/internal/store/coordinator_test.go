package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestCoordinateAndSummaries(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := dbtest.NewDB(t)
	store := New(observation.TestContextTB(t), database.NewDB(logger, db))

	now1 := timeutil.Now().UTC()
	now2 := now1.Add(time.Hour * 2)
	now3 := now2.Add(time.Hour * 5)
	now4 := now2.Add(time.Hour * 7)

	key1 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")
	key2 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "234")
	key3 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "345")

	//
	// Insert data

	testNow = func() time.Time { return now1 }
	if err := store.Coordinate(ctx, key1); err != nil {
		t.Fatalf("unexpected error running coordinate: %s", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE codeintel_ranking_progress SET mapper_completed_at = $1, seed_mapper_completed_at = $1`, now2); err != nil {
		t.Fatalf("unexpected error modifying progress: %s", err)
	}

	testNow = func() time.Time { return now2 }
	if err := store.Coordinate(ctx, key1); err != nil {
		t.Fatalf("unexpected error running coordinate: %s", err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE codeintel_ranking_progress SET reducer_completed_at = $1`, now3); err != nil {
		t.Fatalf("unexpected error modifying progress: %s", err)
	}

	testNow = func() time.Time { return now3 }
	if err := store.Coordinate(ctx, key2); err != nil {
		t.Fatalf("unexpected error running coordinate: %s", err)
	}

	testNow = func() time.Time { return now4 }
	if err := store.Coordinate(ctx, key3); err != nil {
		t.Fatalf("unexpected error running coordinate: %s", err)
	}

	//
	// Gather summaries

	summaries, err := store.Summaries(ctx)
	if err != nil {
		t.Fatalf("unexpected error fetching summaries: %s", err)
	}

	expectedSummaries := []shared.Summary{
		{
			GraphKey:                key3,
			VisibleToZoekt:          false,
			PathMapperProgress:      shared.Progress{StartedAt: now4},
			ReferenceMapperProgress: shared.Progress{StartedAt: now4},
		},
		{
			GraphKey:                key2,
			VisibleToZoekt:          false,
			PathMapperProgress:      shared.Progress{StartedAt: now3},
			ReferenceMapperProgress: shared.Progress{StartedAt: now3},
		},
		{
			GraphKey:                key1,
			VisibleToZoekt:          true,
			PathMapperProgress:      shared.Progress{StartedAt: now1, CompletedAt: &now2},
			ReferenceMapperProgress: shared.Progress{StartedAt: now1, CompletedAt: &now2},
			ReducerProgress:         &shared.Progress{StartedAt: now2, CompletedAt: &now3},
		},
	}
	if diff := cmp.Diff(expectedSummaries, summaries); diff != "" {
		t.Errorf("unexpected summaries (-want +got):\n%s", diff)
	}
}
