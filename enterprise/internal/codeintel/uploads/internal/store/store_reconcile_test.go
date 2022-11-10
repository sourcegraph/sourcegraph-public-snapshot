package store

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestReconcileCandidates(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (102, 50, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (103, 50, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (104, 50, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (105, 50, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}', 'completed');
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Initial batch of records
	ids, err := store.ReconcileCandidates(ctx, 4)
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}
	expectedIDs := []int{
		100,
		101,
		102,
		103,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}

	// Remaining records, wrap around
	ids, err = store.ReconcileCandidates(ctx, 4)
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}
	expectedIDs = []int{
		100,
		101,
		104,
		105,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}
}
