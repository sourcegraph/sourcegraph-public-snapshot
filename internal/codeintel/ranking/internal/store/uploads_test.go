package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetUploadsForRanking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
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
		{ExportedUploadID: 2, UploadID: 102, Repo: "bar", RepoID: 51},
		{ExportedUploadID: 1, UploadID: 103, Repo: "baz", RepoID: 52},
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
		{ExportedUploadID: 3, UploadID: 100, Repo: "foo", RepoID: 50},
	}
	if diff := cmp.Diff(expectedUploads, uploads); diff != "" {
		t.Fatalf("unexpected uploads (-want +got):\n%s", diff)
	}
}

func TestVacuumAbandonedExportedUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	// Insert uploads
	for i := 1; i <= 9; i++ {
		insertUploads(t, db,
			uploadsshared.Upload{ID: 10 + i},
			uploadsshared.Upload{ID: 20 + i},
		)
	}

	// Insert exported uploads
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $2, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRankingGraphKey,
		mockRankingGraphKey+"-old",
	); err != nil {
		t.Fatalf("unexpected error inserting exported upload record: %s", err)
	}

	assertCounts := func(expectedNumRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numExportRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_exports`)))
		if err != nil {
			t.Fatalf("failed to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fatalf("unexpected number of definition records. want=%d have=%d", expectedNumRecords, numExportRecords)
		}
	}

	// assert initial count
	assertCounts(9 + 9)

	_, err := store.VacuumAbandonedExportedUploads(ctx, mockRankingGraphKey, 1000)
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted exported uploads: %s", err)
	}

	// only records associated with key remain
	assertCounts(9)
}

func TestSoftDeleteStaleExportedUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	// Insert uploads
	for i := 1; i <= 9; i++ {
		insertUploads(t, db,
			uploadsshared.Upload{ID: 10 + i},
			uploadsshared.Upload{ID: 20 + i},
		)
	}

	// make uploads 11, 14, 22, and 27 visible at tip of their repo
	insertVisibleAtTip(t, db, 50, 11, 14, 22, 27)

	// Insert exported uploads
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("unexpected error inserting exported upload record: %s", err)
	}

	assertCounts := func(expectedNumRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numExportRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`
			SELECT COUNT(*)
			FROM codeintel_ranking_exports
			WHERE deleted_at IS NULL
		`)))
		if err != nil {
			t.Fatalf("failed to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fatalf("unexpected number of definition records. want=%d have=%d", expectedNumRecords, numExportRecords)
		}
	}

	// assert initial count
	assertCounts(9 + 9)

	_, _, err := store.SoftDeleteStaleExportedUploads(ctx, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error soft-deleting expired uploads: %s", err)
	}

	// only records visible at tip
	assertCounts(4)
}

func TestVacuumDeletedExportedUploads(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	// Insert uploads
	for i := 1; i <= 9; i++ {
		insertUploads(t, db,
			uploadsshared.Upload{ID: 10 + i},
			uploadsshared.Upload{ID: 20 + i},
		)
	}

	// make uploads 11, 14, 22, and 27 visible at tip of their repo
	insertVisibleAtTip(t, db, 50, 11, 14, 22, 27)

	// Insert exported uploads
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("unexpected error inserting exported upload record: %s", err)
	}

	assertCounts := func(expectedNumRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numExportRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_exports`)))
		if err != nil {
			t.Fatalf("failed to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fatalf("unexpected number of definition records. want=%d have=%d", expectedNumRecords, numExportRecords)
		}
	}

	// assert initial count
	assertCounts(9 + 9)

	_, _, err := store.SoftDeleteStaleExportedUploads(ctx, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error soft-deleting expired uploads: %s", err)
	}

	// records only soft-deleted
	assertCounts(9 + 9)

	_, err = store.VacuumDeletedExportedUploads(ctx, rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123"))
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted uploads: %s", err)
	}

	// only non-soft-deleted records remain
	assertCounts(4)
}
