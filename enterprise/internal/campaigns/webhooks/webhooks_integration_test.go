package webhooks

import (
	"testing"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestWebhooksIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)

	user := ct.CreateTestUser(t, false)

	t.Run("GitHubWebhook", testGitHubWebhook(dbconn.Global, user.ID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(dbconn.Global, user.ID))
	t.Run("GitLabWebhook", testGitLabWebhook(dbconn.Global, user.ID))
}
