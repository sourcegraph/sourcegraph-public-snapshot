package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestMarkRepoRevsAsProcessed(t *testing.T) {
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

	// mark first elements as complete; re-request remaining
	if err := store.MarkRepoRevsAsProcessed(ctx, []int{1, 2, 3, 4, 5}); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	repoRevs, err = store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[5:], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-want +got):\n%s", diff)
	}
}
