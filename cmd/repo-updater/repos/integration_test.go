package repos_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
)

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()
	db, cleanup := testDatabase(t)
	defer cleanup()

	store := repos.NewDBStore(ctx, db, sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"DBStore/Transact", testDBStoreTransact(store)},
		{"DBStore/ListExternalServices", testStoreListExternalServices(store)},
		{"DBStore/UpsertExternalServices", testStoreUpsertExternalServices(store)},
		{"DBStore/GetRepoByName", testStoreGetRepoByName(store)},
		{"DBStore/UpsertRepos", testStoreUpsertRepos(store)},
		{"DBStore/ListRepos", testStoreListRepos(store)},
		{"Syncer/Sync", testSyncerSync(store)},
		{"Migrations/GithubSetDefaultRepositoryQuery",
			testGithubSetDefaultRepositoryQueryMigration(store)},
	} {
		t.Run(tc.name, tc.test)
	}
}
