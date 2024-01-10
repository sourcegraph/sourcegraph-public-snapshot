package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetIndexers(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)
	ctx := context.Background()

	insertUploads(t, db,
		shared.Upload{ID: 1, Indexer: "scip-typescript"},
		shared.Upload{ID: 2, Indexer: "scip-typescript"},
		shared.Upload{ID: 3, Indexer: "scip-typescript"},
		shared.Upload{ID: 4, Indexer: "scip-typescript"},
		shared.Upload{ID: 5, Indexer: "scip-typescript"},
		shared.Upload{ID: 6, Indexer: "lsif-ocaml", RepositoryID: 51},
		shared.Upload{ID: 7, Indexer: "lsif-ocaml", RepositoryID: 51},
		shared.Upload{ID: 8, Indexer: "third-party/scip-python@sha256:deadbeefdeadbeefdeadbeef", RepositoryID: 51},
	)

	// Global
	indexers, err := store.GetIndexers(ctx, shared.GetIndexersOptions{})
	if err != nil {
		t.Fatalf("unexpected error getting indexers: %s", err)
	}
	expectedIndexers := []string{
		"lsif-ocaml",
		"scip-typescript",
		"third-party/scip-python@sha256:deadbeefdeadbeefdeadbeef",
	}
	if diff := cmp.Diff(expectedIndexers, indexers); diff != "" {
		t.Errorf("unexpected indexers (-want +got):\n%s", diff)
	}

	// Repo-specific
	indexers, err = store.GetIndexers(ctx, shared.GetIndexersOptions{RepositoryID: 51})
	if err != nil {
		t.Fatalf("unexpected error getting indexers: %s", err)
	}
	expectedIndexers = []string{
		"lsif-ocaml",
		"third-party/scip-python@sha256:deadbeefdeadbeefdeadbeef",
	}
	if diff := cmp.Diff(expectedIndexers, indexers); diff != "" {
		t.Errorf("unexpected indexers (-want +got):\n%s", diff)
	}
}

func TestRecentUploadsSummary(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	t0 := time.Unix(1587396557, 0).UTC()
	t1 := t0.Add(-time.Minute * 1)
	t2 := t0.Add(-time.Minute * 2)
	t3 := t0.Add(-time.Minute * 3)
	t4 := t0.Add(-time.Minute * 4)
	t5 := t0.Add(-time.Minute * 5)
	t6 := t0.Add(-time.Minute * 6)
	t7 := t0.Add(-time.Minute * 7)
	t8 := t0.Add(-time.Minute * 8)
	t9 := t0.Add(-time.Minute * 9)

	r1 := 1
	r2 := 2

	addDefaults := func(upload shared.Upload) shared.Upload {
		upload.Commit = makeCommit(upload.ID)
		upload.RepositoryID = 50
		upload.RepositoryName = "n-50"
		upload.IndexerVersion = "latest"
		upload.UploadedParts = []int{}
		return upload
	}

	uploads := []shared.Upload{
		addDefaults(shared.Upload{ID: 150, UploadedAt: t0, Root: "r1", Indexer: "i1", State: "queued", Rank: &r2}), // visible (group 1)
		addDefaults(shared.Upload{ID: 151, UploadedAt: t1, Root: "r1", Indexer: "i1", State: "queued", Rank: &r1}), // visible (group 1)
		addDefaults(shared.Upload{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", State: "errored"}),          // visible (group 1)
		addDefaults(shared.Upload{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", State: "completed"}),        // visible (group 2)
		addDefaults(shared.Upload{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", State: "completed"}),        // visible (group 3)
		addDefaults(shared.Upload{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", State: "completed"}),        // visible (group 4)
		addDefaults(shared.Upload{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
		addDefaults(shared.Upload{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", State: "errored"}),          // shadowed
	}
	insertUploads(t, db, uploads...)

	summary, err := store.GetRecentUploadsSummary(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying recent upload summary: %s", err)
	}

	expected := []shared.UploadsWithRepositoryNamespace{
		{Root: "r1", Indexer: "i1", Uploads: []shared.Upload{uploads[0], uploads[1], uploads[2]}},
		{Root: "r1", Indexer: "i2", Uploads: []shared.Upload{uploads[3]}},
		{Root: "r2", Indexer: "i1", Uploads: []shared.Upload{uploads[4]}},
		{Root: "r2", Indexer: "i2", Uploads: []shared.Upload{uploads[6]}},
	}
	if diff := cmp.Diff(expected, summary); diff != "" {
		t.Errorf("unexpected upload summary (-want +got):\n%s", diff)
	}
}

func TestRecentIndexesSummary(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	t0 := time.Unix(1587396557, 0).UTC()
	t1 := t0.Add(-time.Minute * 1)
	t2 := t0.Add(-time.Minute * 2)
	t3 := t0.Add(-time.Minute * 3)
	t4 := t0.Add(-time.Minute * 4)
	t5 := t0.Add(-time.Minute * 5)
	t6 := t0.Add(-time.Minute * 6)
	t7 := t0.Add(-time.Minute * 7)
	t8 := t0.Add(-time.Minute * 8)
	t9 := t0.Add(-time.Minute * 9)

	r1 := 1
	r2 := 2

	addDefaults := func(index uploadsshared.Index) uploadsshared.Index {
		index.Commit = makeCommit(index.ID)
		index.RepositoryID = 50
		index.RepositoryName = "n-50"
		index.DockerSteps = []uploadsshared.DockerStep{}
		index.IndexerArgs = []string{}
		index.LocalSteps = []string{}
		return index
	}

	indexes := []uploadsshared.Index{
		addDefaults(uploadsshared.Index{ID: 150, QueuedAt: t0, Root: "r1", Indexer: "i1", State: "queued", Rank: &r2}), // visible (group 1)
		addDefaults(uploadsshared.Index{ID: 151, QueuedAt: t1, Root: "r1", Indexer: "i1", State: "queued", Rank: &r1}), // visible (group 1)
		addDefaults(uploadsshared.Index{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", State: "errored"}),        // visible (group 1)
		addDefaults(uploadsshared.Index{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", State: "completed"}),      // visible (group 2)
		addDefaults(uploadsshared.Index{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", State: "completed"}),      // visible (group 3)
		addDefaults(uploadsshared.Index{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", State: "errored"}),        // shadowed
		addDefaults(uploadsshared.Index{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", State: "completed"}),      // visible (group 4)
		addDefaults(uploadsshared.Index{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
		addDefaults(uploadsshared.Index{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
		addDefaults(uploadsshared.Index{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", State: "errored"}),        // shadowed
	}
	insertIndexes(t, db, indexes...)

	summary, err := store.GetRecentIndexesSummary(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error querying recent index summary: %s", err)
	}

	expected := []uploadsshared.IndexesWithRepositoryNamespace{
		{Root: "r1", Indexer: "i1", Indexes: []uploadsshared.Index{indexes[0], indexes[1], indexes[2]}},
		{Root: "r1", Indexer: "i2", Indexes: []uploadsshared.Index{indexes[3]}},
		{Root: "r2", Indexer: "i1", Indexes: []uploadsshared.Index{indexes[4]}},
		{Root: "r2", Indexer: "i2", Indexes: []uploadsshared.Index{indexes[6]}},
	}
	if diff := cmp.Diff(expected, summary); diff != "" {
		t.Errorf("unexpected index summary (-want +got):\n%s", diff)
	}
}

func TestRepositoryIDsWithErrors(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	now := time.Now()
	t1 := now.Add(-time.Minute * 1)
	t2 := now.Add(-time.Minute * 2)
	t3 := now.Add(-time.Minute * 3)

	insertUploads(t, db,
		shared.Upload{ID: 100, RepositoryID: 50},                  // Repo 50 = success (no index)
		shared.Upload{ID: 101, RepositoryID: 51},                  // Repo 51 = success (+ successful index)
		shared.Upload{ID: 103, RepositoryID: 53, State: "failed"}, // Repo 53 = failed

		// Repo 54 = multiple failures for same project
		shared.Upload{ID: 150, RepositoryID: 54, State: "failed", FinishedAt: &t1},
		shared.Upload{ID: 151, RepositoryID: 54, State: "failed", FinishedAt: &t2},
		shared.Upload{ID: 152, RepositoryID: 54, State: "failed", FinishedAt: &t3},

		// Repo 55 = multiple failures for different projects
		shared.Upload{ID: 160, RepositoryID: 55, State: "failed", FinishedAt: &t1, Root: "proj1"},
		shared.Upload{ID: 161, RepositoryID: 55, State: "failed", FinishedAt: &t2, Root: "proj2"},
		shared.Upload{ID: 162, RepositoryID: 55, State: "failed", FinishedAt: &t3, Root: "proj3"},

		// Repo 58 = multiple failures with later success (not counted)
		shared.Upload{ID: 170, RepositoryID: 58, State: "completed", FinishedAt: &t1},
		shared.Upload{ID: 171, RepositoryID: 58, State: "failed", FinishedAt: &t2},
		shared.Upload{ID: 172, RepositoryID: 58, State: "failed", FinishedAt: &t3},
	)
	insertIndexes(t, db,
		uploadsshared.Index{ID: 201, RepositoryID: 51},                  // Repo 51 = success
		uploadsshared.Index{ID: 202, RepositoryID: 52, State: "failed"}, // Repo 52 = failing index
		uploadsshared.Index{ID: 203, RepositoryID: 53},                  // Repo 53 = success (+ failing upload)

		// Repo 56 = multiple failures for same project
		uploadsshared.Index{ID: 250, RepositoryID: 56, State: "failed", FinishedAt: &t1},
		uploadsshared.Index{ID: 251, RepositoryID: 56, State: "failed", FinishedAt: &t2},
		uploadsshared.Index{ID: 252, RepositoryID: 56, State: "failed", FinishedAt: &t3},

		// Repo 57 = multiple failures for different projects
		uploadsshared.Index{ID: 260, RepositoryID: 57, State: "failed", FinishedAt: &t1, Root: "proj1"},
		uploadsshared.Index{ID: 261, RepositoryID: 57, State: "failed", FinishedAt: &t2, Root: "proj2"},
		uploadsshared.Index{ID: 262, RepositoryID: 57, State: "failed", FinishedAt: &t3, Root: "proj3"},
	)

	// Query page 1
	repositoriesWithCount, totalCount, err := store.RepositoryIDsWithErrors(ctx, 0, 4)
	if err != nil {
		t.Fatalf("unexpected error getting repositories with errors: %s", err)
	}
	if expected := 6; totalCount != expected {
		t.Fatalf("unexpected total number of repositories. want=%d have=%d", expected, totalCount)
	}
	expected := []uploadsshared.RepositoryWithCount{
		{RepositoryID: 55, Count: 3},
		{RepositoryID: 57, Count: 3},
		{RepositoryID: 52, Count: 1},
		{RepositoryID: 53, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}

	// Query page 2
	repositoriesWithCount, _, err = store.RepositoryIDsWithErrors(ctx, 4, 4)
	if err != nil {
		t.Fatalf("unexpected error getting repositories with errors: %s", err)
	}
	expected = []uploadsshared.RepositoryWithCount{
		{RepositoryID: 54, Count: 1},
		{RepositoryID: 56, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-want +got):\n%s", diff)
	}
}

func TestNumRepositoriesWithCodeIntelligence(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(logger, sqlDB)
	store := New(&observation.TestContext, db)

	insertUploads(t, db,
		shared.Upload{ID: 100, RepositoryID: 50},
		shared.Upload{ID: 101, RepositoryID: 51},
		shared.Upload{ID: 102, RepositoryID: 52}, // Not in commit graph
		shared.Upload{ID: 103, RepositoryID: 53}, // Not on default branch
	)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO lsif_uploads_visible_at_tip
			(repository_id, upload_id, is_default_branch)
		VALUES
			(50, 100, true),
			(51, 101, true),
			(53, 103, false)
	`); err != nil {
		t.Fatalf("unexpected error inserting visible uploads: %s", err)
	}

	count, err := store.NumRepositoriesWithCodeIntelligence(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting top repositories to configure: %s", err)
	}
	if expected := 2; count != expected {
		t.Fatalf("unexpected number of repositories. want=%d have=%d", expected, count)
	}
}
