package store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
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
		`INSERT INTO repo (id, name, deleted_at, private) VALUES (%s, %s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
		false,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}
}

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
