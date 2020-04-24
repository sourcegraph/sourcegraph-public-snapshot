package campaigns

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

	t.Run("Store", testStore(dbtest.NewDB(t, *dsn)))

	// The following tests need to be separate because testStore above wraps everything in a global transaction
	t.Run("GitHubWebhook", testGitHubWebhook(dbtest.NewDB(t, *dsn)))
	t.Run("StoreLocking", testStoreLocking(dbtest.NewDB(t, *dsn)))
	t.Run("ProcessChangesetJob", testProcessChangesetJob(dbtest.NewDB(t, *dsn)))
}
