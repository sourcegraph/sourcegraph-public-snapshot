package graphs

import (
	"database/sql"
	"flag"
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

	t.Run("Store", func(t *testing.T) {
		t.Run("Graphs", storeTest(db, testStoreGraphs))
	})
}

func insertTestOrg(t *testing.T, db *sql.DB) (orgID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO orgs (name) VALUES ('graphs-org') RETURNING id").Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}

	return orgID
}

func insertTestUser(t *testing.T, db *sql.DB) (userID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO users (username) VALUES ('graphs-user') RETURNING id").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}
