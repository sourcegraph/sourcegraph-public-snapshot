package webhooks

import (
	"testing"

	"github.com/sourcegraph/log/logtest"

	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestWebhooksIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(logger, sqlDB)

	user := bt.CreateTestUser(t, db, false)

	t.Run("GitHubWebhook", testGitHubWebhook(db, user.ID))
	t.Run("BitbucketServerWebhook", testBitbucketServerWebhook(db, user.ID))
	t.Run("GitLabWebhook", testGitLabWebhook(sqlDB))
	t.Run("BitbucketCloudWebhook", testBitbucketCloudWebhook(sqlDB))
}
