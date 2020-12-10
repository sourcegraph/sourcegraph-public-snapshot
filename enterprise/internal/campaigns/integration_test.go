package campaigns

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)

	user := createTestUser(t, false)

	t.Run("Store", func(t *testing.T) {
		t.Run("Campaigns", storeTest(dbconn.Global, testStoreCampaigns))
		t.Run("Changesets", storeTest(dbconn.Global, testStoreChangesets))
		t.Run("ChangesetEvents", storeTest(dbconn.Global, testStoreChangesetEvents))
		t.Run("ListChangesetSyncData", storeTest(dbconn.Global, testStoreListChangesetSyncData))
		t.Run("ListChangesetsTextSearch", storeTest(dbconn.Global, testStoreListChangesetsTextSearch))
		t.Run("CampaignSpecs", storeTest(dbconn.Global, testStoreCampaignSpecs))
		t.Run("ChangesetSpecs", storeTest(dbconn.Global, testStoreChangesetSpecs))
		t.Run("CodeHosts", storeTest(dbconn.Global, testStoreCodeHost))
	})

	t.Run("GitHubWebhook", testGitHubWebhook(dbconn.Global, user.ID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(dbconn.Global, user.ID))
	t.Run("GitLabWebhook", testGitLabWebhook(dbconn.Global, user.ID))
}
