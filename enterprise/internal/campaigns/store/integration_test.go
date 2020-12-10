package store

import (
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	dbtesting.SetupGlobalTestDB(t)

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
}

func createTestUser(t *testing.T) *types.User {
	t.Helper()
	user := &types.User{
		Username:    "testuser-1",
		DisplayName: "testuser",
	}
	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", user.Username, false)
	err := dbconn.Global.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbconn.Global.Exec("INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
	if err != nil {
		t.Fatalf("failed to create name: %s", err)
	}
	return user
}
