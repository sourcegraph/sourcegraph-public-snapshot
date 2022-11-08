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
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (102, 51, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (103, 52, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (104, 52, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (105, 53, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (100, 50, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (102, 51, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (103, 52, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Initial batch of records
	uploads, err := store.GetUploadsForRanking(ctx, "test", "ranking", 2)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads := []ExportedUpload{
		{ID: 102, Repo: "bar", ObjectPrefix: "ranking/test/102"},
		{ID: 103, Repo: "baz", ObjectPrefix: "ranking/test/103"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}

	// Remaining records
	uploads, err = store.GetUploadsForRanking(ctx, "test", "ranking", 2)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads = []ExportedUpload{
		{ID: 100, Repo: "foo", ObjectPrefix: "ranking/test/100"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}
}

func TestProcessStaleExportedUploads(t *testing.T) {
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
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (102, 50, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (103, 51, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (104, 51, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads (id, repository_id, commit, indexer, num_parts, uploaded_parts, state) VALUES (105, 52, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (100, 50, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (103, 51, true);
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (105, 52, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Insert all records
	uploads, err := store.GetUploadsForRanking(ctx, "test", "ranking", 10)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads := []ExportedUpload{
		{ID: 100, Repo: "foo", ObjectPrefix: "ranking/test/100"}, // shadowed by upload 102
		{ID: 103, Repo: "bar", ObjectPrefix: "ranking/test/103"}, // repo gets deleted
		{ID: 105, Repo: "baz", ObjectPrefix: "ranking/test/105"}, // upload gets deleted
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}

	// Mess some stuff up
	if _, err := db.ExecContext(ctx, `
		UPDATE repo SET deleted_at = NOW() WHERE id = 51;              -- delete repo (attached to upload 103)
		DELETE FROM lsif_uploads_visible_at_tip WHERE upload_id = 103; -- eventual effect (after janitor runs)
		DELETE FROM lsif_uploads WHERE id = 105;                       -- delete upload
		DELETE FROM lsif_uploads_visible_at_tip WHERE upload_id = 105; -- eventual effect (after janitor runs)

		DELETE FROM lsif_uploads_visible_at_tip WHERE upload_id = 100; -- Shadow upload 100 with upload 102
		INSERT INTO lsif_uploads_visible_at_tip (upload_id, repository_id, is_default_branch) VALUES (102, 50, true);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Assert that these records will be marked for deletion
	var deletedObjectPrefixes []string
	numDeleted, err := store.ProcessStaleExportedUploads(ctx, "test", 100, func(ctx context.Context, objectPrefix string) error {
		deletedObjectPrefixes = append(deletedObjectPrefixes, objectPrefix)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error processing stale exported uploads: %s", err)
	}
	if numDeleted != len(deletedObjectPrefixes) {
		t.Fatalf("expected numDeleted to match number of invocations. numDeleted=%d len(deleted)=%d", numDeleted, len(deletedObjectPrefixes))
	}

	expectedDeletedObjectPrefixes := []string{
		"ranking/test/100",
		"ranking/test/103",
		"ranking/test/105",
	}
	sort.Strings(deletedObjectPrefixes)
	if diff := cmp.Diff(expectedDeletedObjectPrefixes, deletedObjectPrefixes); diff != "" {
		t.Fatalf("unexpected deleted IDs (-want +got):\n%s", diff)
	}
}
