package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertDependencyIndexingJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	insertRepo(t, db, 50, "")

	insertUploads(t, db, upload{
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

func TestGetQueuedRepoRev(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	expected := []RepoRev{
		{1, 50, "HEAD"},
		{2, 50, "HEAD~1"},
		{3, 50, "HEAD~2"},
		{4, 51, "HEAD"},
		{5, 51, "HEAD~1"},
		{6, 51, "HEAD~2"},
		{7, 52, "HEAD"},
		{8, 52, "HEAD~1"},
		{9, 52, "HEAD~2"},
	}
	for _, repoRev := range expected {
		if err := store.QueueRepoRev(ctx, repoRev.RepositoryID, repoRev.Rev); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	}

	// entire set
	repoRevs, err := store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}

	// smaller page size
	repoRevs, err = store.GetQueuedRepoRev(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[:5], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}
}
