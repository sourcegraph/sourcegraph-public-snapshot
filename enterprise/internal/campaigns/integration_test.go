package campaigns

import (
	"context"
	"database/sql"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db := dbtest.NewDB(t, *dsn)

	userID := insertTestUser(t, db)

	t.Run("Store", func(t *testing.T) {
		t.Run("Campaigns", storeTest(db, testCampaigns))
		t.Run("Changesets", storeTest(db, testChangesets))
		t.Run("ChangesetEvents", testChangesetEvents(db))
		t.Run("ListChangesetSyncData", testListChangesetSyncData(db))
		t.Run("PatchSets", testPatchSets(db))
		t.Run("PatchSets_DeleteExpired", testPatchSetsDeleteExpired(db))
		t.Run("Patches", testPatches(db))
		t.Run("ChangesetJobs", testChangesetJobs(db))
	})

	t.Run("GitHubWebhook", testGitHubWebhook(db, userID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(db, userID))
	t.Run("MigratePatchesWithoutDiffStats", testMigratePatchesWithoutDiffStats(db, userID))

	// The following tests need to be separate because testStore above wraps everything in a global transaction
	t.Run("StoreLocking", testStoreLocking(db))
	t.Run("ProcessChangesetJob", testProcessChangesetJob(db, userID))
}

func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}

func insertTestUser(t *testing.T, db *sql.DB) (userID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO users (username) VALUES ('bbs-admin') RETURNING id").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

type clock interface {
	now() time.Time
	add(time.Duration) time.Time
}

type testClock struct {
	t time.Time
}

func (c *testClock) now() time.Time                { return c.t }
func (c *testClock) add(d time.Duration) time.Time { c.t = c.t.Add(d); return c.t }

type storeTestFunc func(*testing.T, context.Context, *Store, repos.Store, clock)

func storeTest(db *sql.DB, f storeTestFunc) func(*testing.T) {
	return func(t *testing.T) {
		c := &testClock{t: time.Now().UTC().Truncate(time.Microsecond)}

		// Store tests all run in a transaction that's rolled back at the end
		// of the tests, so that foreign key constraints can be deferred and we
		// don't need to insert a lot of dependencies into the DB (users,
		// repos, ...) to setup the tests.
		tx := dbtest.NewTx(t, db)
		s := NewStoreWithClock(tx, c.now)

		rs := repos.NewDBStore(db, sql.TxOptions{})

		f(t, context.Background(), s, rs, c)
	}
}
