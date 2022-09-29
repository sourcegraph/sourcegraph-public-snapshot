package dbstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestRepoNames(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)
	ctx := context.Background()

	insertRepo(t, db, 50, "A")
	insertRepo(t, db, 51, "B")
	insertRepo(t, db, 52, "C")
	insertRepo(t, db, 53, "D")
	insertRepo(t, db, 54, "E")
	insertRepo(t, db, 55, "F")

	names, err := store.RepoNames(ctx, 50, 52, 53, 54, 57)
	if err != nil {
		t.Fatalf("unexpected error querying repository names: %s", err)
	}

	expected := map[int]string{
		50: "A",
		52: "C",
		53: "D",
		54: "E",
	}
	if diff := cmp.Diff(expected, names); diff != "" {
		t.Errorf("unexpected repository names (-want +got):\n%s", diff)
	}
}
