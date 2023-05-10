package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
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
	store := New(&observation.TestContext, db)

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
	expectedUploads := []uploadsshared.ExportedUpload{
		{ID: 102, Repo: "bar", RepoID: 51, ObjectPrefix: "ranking/test/102"},
		{ID: 103, Repo: "baz", RepoID: 52, ObjectPrefix: "ranking/test/103"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}

	// Remaining records
	uploads, err = store.GetUploadsForRanking(ctx, "test", "ranking", 2)
	if err != nil {
		t.Fatalf("unexpected error getting uploads for ranking: %s", err)
	}
	expectedUploads = []uploadsshared.ExportedUpload{
		{ID: 100, Repo: "foo", RepoID: 50, ObjectPrefix: "ranking/test/100"},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}
}

func TestVacuumAbandonedExportedUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// TODO - setup

	_, err := store.VacuumAbandonedExportedUploads(ctx, mockRankingGraphKey, 100)
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted exported uploads: %s", err)
	}

	// TODO - assertions
}
