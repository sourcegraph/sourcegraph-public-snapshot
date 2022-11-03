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

func TestGetUploadsForRanking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (id, name, deleted_at) VALUES (50, 'foo', NULL);
		INSERT INTO repo (id, name, deleted_at) VALUES (51, 'bar', NULL);
		INSERT INTO repo (id, name, deleted_at) VALUES (52, 'baz', NULL);
		INSERT INTO repo (id, name, deleted_at) VALUES (53, 'del', NOW());
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (102, 51, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (103, 52, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (104, 52, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (105, 53, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (100, 50, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (102, 51, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (103, 52, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Initial batch of records
	uploads, err := store.GetUploadsForRanking(ctx, "test", 2)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads := []Upload{
		{ID: 102, Repo: "bar"},
		{ID: 103, Repo: "baz"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}

	// Remaining records
	uploads, err = store.GetUploadsForRanking(ctx, "test", 2)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads = []Upload{
		{ID: 100, Repo: "foo"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}
}

func TestProcessStaleExportedUplods(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (id, name, deleted_at) VALUES (50, 'foo', NULL);
		INSERT INTO repo (id, name, deleted_at) VALUES (51, 'bar', NULL);
		INSERT INTO repo (id, name, deleted_at) VALUES (52, 'baz', NULL);
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (102, 50, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (103, 51, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (104, 51, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts) VALUES (105, 52, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}');
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (100, 50, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (103, 51, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (105, 52, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Insert all records
	uploads, err := store.GetUploadsForRanking(ctx, "test", 10)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads := []Upload{
		{ID: 100, Repo: "foo"}, // replaced by upload 102
		{ID: 103, Repo: "bar"}, // repo gets deleted
		{ID: 105, Repo: "baz"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}

	// Mess some stuff up
	if _, err := db.ExecContext(ctx, `
		UPDATE repo SET deleted_at = NOW() WHERE id = 51;
		-- DELETE FROM lsif_uploads WHERE id = 105;
		DELETE FROM lsif_uploads_visible_at_tip WHERE upload_id = 100;
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (102, 50, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Assert that these records will be marked for deletion
	var deletedIDs []int
	numDeleted, err := store.ProcessStaleExportedUplods(ctx, "test", 100, func(ctx context.Context, id int) error {
		deletedIDs = append(deletedIDs, id)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error processing stale exported uploads: %s", err)
	}
	if numDeleted != len(deletedIDs) {
		t.Fatalf("expected numDeleted to match number of invocations. numDeleted=%d len(deleted)=%d", numDeleted, len(deletedIDs))
	}

	expectedDeletedIDs := []int{
		100,
		103,
		// 105,
	}
	sort.Ints(deletedIDs)
	if diff := cmp.Diff(expectedDeletedIDs, deletedIDs); diff != "" {
		t.Fatalf("unexpected deleted IDs (-want +got):\n%s", diff)
	}
}
