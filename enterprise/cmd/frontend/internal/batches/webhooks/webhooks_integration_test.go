package webhooks

import (
	"testing"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestWebhooksIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)

	user := ct.CreateTestUser(t, db, false)

	t.Run("GitHubWebhook", testGitHubWebhook(sqlDB, user.ID))
	t.Run("BitbucketServerWebhook", testBitbucketServerWebhook(sqlDB, user.ID))
	t.Run("GitLabWebhook", testGitLabWebhook(sqlDB))
	t.Run("BitbucketCloudWebhook", testBitbucketCloudWebhook(sqlDB, user.ID))
}
