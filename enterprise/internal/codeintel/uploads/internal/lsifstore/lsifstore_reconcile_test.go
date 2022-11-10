package lsifstore

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestIDsWithMeta(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t))
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (100, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (102, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (104, 0);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	candidates := []int{
		100, // exists
		101,
		103,
		104, // exists
		105,
	}
	ids, err := store.IDsWithMeta(ctx, candidates)
	if err != nil {
		t.Fatalf("failed to find upload IDs with metadata: %s", err)
	}
	expectedIDs := []int{
		100,
		104,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}
}

func TestReconcileCandidates(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t))
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (100, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (101, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (102, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (103, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (104, 0);
		INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (105, 0);
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
