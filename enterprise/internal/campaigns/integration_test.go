package campaigns

import (
	"database/sql"
	"flag"
	"strings"
	"testing"

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

	t.Run("Store", testStore(db))
	t.Run("GitHubWebhook", testGitHubWebhook(db, userID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(db, userID))

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
