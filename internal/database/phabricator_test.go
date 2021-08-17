package database

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestPhabricatorStore_GetByName(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		ctx   = context.Background()
		db    = dbtest.NewDB(t, "")
		store = Phabricator(db)
		clock = timeutil.NewFakeClock(time.Now(), 0)
		now   = clock.Now()
	)

	// Create a repository that is not soft deleted.
	{
		q := sqlf.Sprintf(`
INSERT INTO phabricator_repos (id, callsign, repo_name, url, created_at, updated_at, deleted_at)
VALUES (%d, %s, %s, %s, %s, %s, %s)
`, 1, "ACTIVE", "phabricator.example.com/active", "https://phabricator.example.com", now, now, nil)
		_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a repository that is soft deleted.
	{
		q := sqlf.Sprintf(`
INSERT INTO phabricator_repos (id, callsign, repo_name, url, created_at, updated_at, deleted_at)
VALUES (%d, %s, %s, %s, %s, %s, %s)
`, 2, "DELETED", "phabricator.example.com/deleted", "https://phabricator.example.com", now, now, now)
		_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Look for the active repository.
	{
		repo, err := store.GetByName(ctx, "phabricator.example.com/active")
		if err != nil {
			t.Fatal(err)
		}
		if repo == nil {
			t.Fatal("could not find active phabricator repository")
		}
	}

	// Look for the deleted repository.
	{
		repo, err := store.GetByName(ctx, "phabricator.example.com/deleted")
		if err != nil && !errcode.IsNotFound(err) {
			t.Fatal(err)
		}
		if repo != nil {
			t.Fatal("returned deleted phabricator repository")
		}
	}
}
