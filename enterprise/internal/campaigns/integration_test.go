package campaigns

import (
	"context"
	"database/sql"
	"flag"
	"strings"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)
	db := dbtest.NewDB(t, *dsn)

	userID := insertTestUser(t, context.Background(), db, "bbs-admin")

	t.Run("Store", func(t *testing.T) {
		t.Run("Campaigns", storeTest(db, testStoreCampaigns))
		t.Run("Changesets", storeTest(db, testStoreChangesets))
		t.Run("ChangesetEvents", storeTest(db, testStoreChangesetEvents))
		t.Run("ListChangesetSyncData", storeTest(db, testStoreListChangesetSyncData))
		t.Run("CampaignSpecs", storeTest(db, testStoreCampaignSpecs))
		t.Run("ChangesetSpecs", storeTest(db, testStoreChangesetSpecs))
		t.Run("UserToken", storeTest(db, testStoreUserTokens))
	})

	t.Run("GitHubWebhook", testGitHubWebhook(db, userID))
	t.Run("BitbucketWebhook", testBitbucketWebhook(db, userID))
	t.Run("GitLabWebhook", testGitLabWebhook(db, userID))
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

func insertTestUser(t *testing.T, ctx context.Context, db dbutil.DB, name string) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf(
		"INSERT INTO users (username) VALUES (%s) RETURNING id",
		name,
	)

	err := db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), name).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}
