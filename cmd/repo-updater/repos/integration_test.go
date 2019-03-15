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

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"DBStore/Transact", testDBStoreTransact(db)},
		{"DBStore/ListExternalServices", testDBStoreListExternalServices(db)},
		{"DBStore/UpsertExternalServices", testDBStoreUpsertExternalServices(db)},
		{"DBStore/GetRepoByName", testDBStoreGetRepoByName(db)},
		{"DBStore/UpsertRepos", testDBStoreUpsertRepos(db)},
		{"DBStore/ListRepos", testDBStoreListRepos(db)},
		{"Syncer/Sync", testSyncerSync(
			repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable}),
		)},
	} {
		t.Run(tc.name, tc.test)
	}
}
