package a8n

import (
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

	db, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	t.Run("Store", testStore(db))
	// This needs to be in its own test because testStore above wraps everything in a transaction
	// which means we are always able to acquire a lock
	t.Run("StoreLocking", testStoreLocking(db))
	t.Run("GitHubWebhook", testGitHubWebhook(db))
}
