package gitserver

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// insertRepo creates a repository record with the given id and name. If there is already a repository
// with the given identifier, nothing happens
func insertRepo(t testing.TB, db database.DB, id int, name string) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HasPrefix(name, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}
}

func TestRepoNames(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := newWithDB(db)
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
