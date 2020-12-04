package campaigns

import (
	"database/sql"
	"flag"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)

	userID := insertTestUser(t, dbconn.Global)

	t.Run("Store", func(t *testing.T) {
		t.Run("Campaigns", storeTest(dbconn.Global, testStoreCampaigns))
		t.Run("Changesets", storeTest(dbconn.Global, testStoreChangesets))
		t.Run("ChangesetEvents", storeTest(dbconn.Global, testStoreChangesetEvents))
		t.Run("ListChangesetSyncData", storeTest(dbconn.Global, testStoreListChangesetSyncData))
		t.Run("CampaignSpecs", storeTest(dbconn.Global, testStoreCampaignSpecs))
		t.Run("ChangesetSpecs", storeTest(dbconn.Global, testStoreChangesetSpecs))
		t.Run("CodeHosts", storeTest(dbconn.Global, testStoreCodeHost))
	})

	t.Run("GitHubWebhook", testGitHubWebhook(dbconn.Global, userID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(dbconn.Global, userID))
	t.Run("GitLabWebhook", testGitLabWebhook(dbconn.Global, userID))
}

func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}

func insertTestOrg(t *testing.T, db *sql.DB) (orgID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO orgs (name) VALUES ('bbs-org') RETURNING id").Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}

	return orgID
}

func insertTestUser(t *testing.T, db *sql.DB) (userID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO users (username) VALUES ('bbs-admin') RETURNING id").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}
