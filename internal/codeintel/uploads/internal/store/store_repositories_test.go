package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSetRepositoryAsDirty(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	tx := basestore.NewWithHandle(db.Handle())

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	for _, repositoryID := range []int{50, 51, 52, 51, 52} {
		if err := store.SetRepositoryAsDirty(context.Background(), repositoryID, tx); err != nil {
			t.Errorf("unexpected error marking repository as dirty: %s", err)
		}
	}

	repositoryIDs, err := store.GetDirtyRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}

	var keys []int
	for repositoryID := range repositoryIDs {
		keys = append(keys, repositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{50, 51, 52}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-want +got):\n%s", diff)
	}
}

func TestGetRepositoriesMaxStaleAge(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for _, id := range []int{50, 51, 52} {
		insertRepo(t, db, id, "")
	}

	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO lsif_dirty_repositories (
			repository_id,
			update_token,
			dirty_token,
			set_dirty_at
		)
		VALUES
			(50, 10, 10, NOW() - '45 minutes'::interval), -- not dirty
			(51, 20, 25, NOW() - '30 minutes'::interval), -- dirty
			(52, 30, 35, NOW() - '20 minutes'::interval), -- dirty
			(53, 40, 45, NOW() - '30 minutes'::interval); -- no associated repo
	`); err != nil {
		t.Fatalf("unexpected error marking repostiory as dirty: %s", err)
	}

	age, err := store.GetRepositoriesMaxStaleAge(context.Background())
	if err != nil {
		t.Fatalf("unexpected error listing dirty repositories: %s", err)
	}
	if age.Round(time.Second) != 30*time.Minute {
		t.Fatalf("unexpected max age. want=%s have=%s", 30*time.Minute, age)
	}
}
