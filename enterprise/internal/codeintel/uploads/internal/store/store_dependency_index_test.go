package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertDependencySyncingJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	uploadID := 42
	insertRepo(t, db, 50, "")
	insertUploads(t, db, types.Upload{
		ID:            uploadID,
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})

	// No error if upload exists
	if _, err := store.InsertDependencySyncingJob(context.Background(), uploadID); err != nil {
		t.Fatalf("unexpected error enqueueing dependency indexing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencySyncingJob(context.Background(), uploadID+1); err == nil {
		t.Fatalf("expected error enqueueing dependency indexing job for unknown upload")
	}
}
