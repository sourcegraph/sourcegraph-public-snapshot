package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestInsertDependencySyncingJob(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertRepo(t, db, 50, "")

	uploadID, err := store.InsertUpload(context.Background(), Upload{
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	// No error if upload exists
	if _, err := store.InsertDependencySyncingJob(context.Background(), uploadID); err != nil {
		t.Fatalf("unexpected error enqueueing dependency indexing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencySyncingJob(context.Background(), uploadID+1); err == nil {
		t.Fatalf("expected error enqueueing dependency indexing job for unknown upload")
	}
}

func TestInsertDependencyIndexingJob(t *testing.T) {
	db := dbtest.NewDB(t)
	store := testStore(db)

	insertRepo(t, db, 50, "")

	uploadID, err := store.InsertUpload(context.Background(), Upload{
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	// No error if upload exists
	if _, err := store.InsertDependencyIndexingJob(context.Background(), uploadID, "asdf", time.Now()); err != nil {
		t.Fatalf("unexpected error enqueueing dependency index queueing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencyIndexingJob(context.Background(), uploadID+1, "asdf", time.Now()); err == nil {
		t.Fatalf("expected error enqueueing dependency index queueing job for unknown upload")
	}
}
