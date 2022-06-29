package syncwebhooks

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSyncWebhooksIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)

	t.Run("GithubSyncHooks", testGitHubSyncHooks(db, 12345))
}
