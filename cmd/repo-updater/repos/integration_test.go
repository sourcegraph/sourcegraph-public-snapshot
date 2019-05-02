package repos_test

import (
	"context"
	"database/sql"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
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

	dbstore := repos.NewDBStore(ctx, db, sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	lg := log15.New()
	lg.SetHandler(log15.DiscardHandler())

	store := repos.NewObservedStore(
		dbstore,
		lg,
		repos.NewStoreMetrics(),
		trace.Tracer{Tracer: opentracing.GlobalTracer()},
	)

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"DBStore/Transact", testDBStoreTransact(dbstore)},
		{"DBStore/ListExternalServices", testStoreListExternalServices(store)},
		{"DBStore/UpsertExternalServices", testStoreUpsertExternalServices(store)},
		{"DBStore/UpsertRepos", testStoreUpsertRepos(store)},
		{"DBStore/ListRepos", testStoreListRepos(store)},
		{"DBStore/ListRepos/Pagination", testStoreListReposPagination(store)},
		{"DBStore/Syncer/Sync", testSyncerSync(store)},
		{"Migrations/GithubSetDefaultRepositoryQuery",
			testGithubSetDefaultRepositoryQueryMigration(store)},
		{"Migrations/GitLabSetDefaultProjectQuery",
			testGitLabSetDefaultProjectQueryMigration(store)},
		{"Migrations/BitbucketServerSetDefaultRepositoryQuery",
			testBitbucketServerSetDefaultRepositoryQueryMigration(store)},
		{"Migrations/BitbucketServerUsername",
			testBitbucketServerUsernameMigration(store)},
		{"Migrations/EnabledStateDeprecationMigration",
			testEnabledStateDeprecationMigration(store)},
	} {
		t.Run(tc.name, tc.test)
	}
}
