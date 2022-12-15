package store

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertDependencyIndexingJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertRepo(t, db, 50, "")

	insertUploads(t, db, Upload{
		ID:            42,
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})

	// No error if upload exists
	if _, err := store.InsertDependencyIndexingJob(context.Background(), 42, "asdf", time.Now()); err != nil {
		t.Fatalf("unexpected error enqueueing dependency index queueing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencyIndexingJob(context.Background(), 43, "asdf", time.Now()); err == nil {
		t.Fatalf("expected error enqueueing dependency index queueing job for unknown upload")
	}
}
