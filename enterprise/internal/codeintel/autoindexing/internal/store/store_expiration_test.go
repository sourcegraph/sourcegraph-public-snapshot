package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestExpireFailedRecords(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	ctx := context.Background()
	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		// young failures (none removed)
		types.Index{ID: 1, RepositoryID: 50, Commit: makeCommit(1), FinishedAt: timePtr(now.Add(-time.Minute * 10)), State: "failed"},
		types.Index{ID: 2, RepositoryID: 50, Commit: makeCommit(2), FinishedAt: timePtr(now.Add(-time.Minute * 20)), State: "failed"},
		types.Index{ID: 3, RepositoryID: 50, Commit: makeCommit(3), FinishedAt: timePtr(now.Add(-time.Minute * 20)), State: "failed"},

		// failures prior to a success (both removed)
		types.Index{ID: 4, RepositoryID: 50, Commit: makeCommit(4), FinishedAt: timePtr(now.Add(-time.Hour * 10)), Root: "foo", State: "completed"},
		types.Index{ID: 5, RepositoryID: 50, Commit: makeCommit(5), FinishedAt: timePtr(now.Add(-time.Hour * 12)), Root: "foo", State: "failed"},
		types.Index{ID: 6, RepositoryID: 50, Commit: makeCommit(6), FinishedAt: timePtr(now.Add(-time.Hour * 14)), Root: "foo", State: "failed"},

		// old failures (one is left for debugging)
		types.Index{ID: 7, RepositoryID: 51, Commit: makeCommit(7), FinishedAt: timePtr(now.Add(-time.Hour * 3)), State: "failed"},
		types.Index{ID: 8, RepositoryID: 51, Commit: makeCommit(8), FinishedAt: timePtr(now.Add(-time.Hour * 4)), State: "failed"},
		types.Index{ID: 9, RepositoryID: 51, Commit: makeCommit(9), FinishedAt: timePtr(now.Add(-time.Hour * 5)), State: "failed"},

		// failures prior to queued uploads (one removed; queued does not reset failures)
		types.Index{ID: 10, RepositoryID: 52, Commit: makeCommit(10), Root: "foo", State: "queued"},
		types.Index{ID: 11, RepositoryID: 52, Commit: makeCommit(11), FinishedAt: timePtr(now.Add(-time.Hour * 12)), Root: "foo", State: "failed"},
		types.Index{ID: 12, RepositoryID: 52, Commit: makeCommit(12), FinishedAt: timePtr(now.Add(-time.Hour * 14)), Root: "foo", State: "failed"},
	)

	if err := store.ExpireFailedRecords(ctx, 100, time.Hour, now); err != nil {
		t.Fatalf("unexpected error expiring failed records: %s", err)
	}

	ids, err := basestore.ScanInts(db.QueryContext(ctx, "SELECT id FROM lsif_indexes"))
	if err != nil {
		t.Fatalf("unexpected error fetching index ids: %s", err)
	}

	expectedIDs := []int{
		1, 2, 3, // none deleted
		4,      // 5, 6 deleted
		7,      // 8, 9 deleted
		10, 11, // 12 deleted
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
