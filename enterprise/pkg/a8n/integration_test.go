package a8n

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	tx, done := dbtest.NewTx(t, db)
	defer done()

	dbtx := dbutil.NewDBTx(tx)

	t.Run("Store", testStore(dbtx))
	t.Run("GitHubWebhook", testGitHubWebhook(dbtx))
}
