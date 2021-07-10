package repos_test

import (
	"database/sql"
	"flag"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	for _, tc := range []struct {
		name string
		test func(*repos.Store) func(*testing.T)
	}{
		{"DBStore/SyncRateLimiters", testSyncRateLimiters},
		{"DBStore/UpsertRepos", testStoreUpsertRepos},
		{"DBStore/UpsertSources", testStoreUpsertSources},
		{"DBStore/EnqueueSyncJobs", testStoreEnqueueSyncJobs},
		{"DBStore/EnqueueSingleSyncJob", testStoreEnqueueSingleSyncJob},
		{"DBStore/ListExternalRepoSpecs", testStoreListExternalRepoSpecs},
		{"DBStore/SetClonedRepos", testStoreSetClonedRepos},
		{"DBStore/CountNotClonedRepos", testStoreCountNotClonedRepos},
		{"DBStore/Syncer/Sync", testSyncerSync},
		{"DBStore/Syncer/SyncRepo", testSyncRepo},
		{"DBStore/Syncer/SyncWorker", testSyncWorkerPlumbing},
		{"DBStore/Syncer/Run", testSyncRun},
		{"DBStore/Syncer/MultipleServices", testSyncer},
		{"DBStore/Syncer/OrphanedRepos", testOrphanedRepo},
		{"DBStore/Syncer/UserAddedRepos", testUserAddedRepos},
		{"DBStore/Syncer/DeleteExternalService", testDeleteExternalService},
		{"DBStore/Syncer/NameConflict", testNameConflict},
		{"DBStore/Syncer/ConflictingSyncers", testConflictingSyncers},
		{"DBStore/Syncer/SyncRepoMaintainsOtherSources", testSyncRepoMaintainsOtherSources},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := dbtest.NewDB(t, *dsn)

			store := repos.NewStore(db, sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			})

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			store.Log = lg
			store.Metrics = repos.NewStoreMetrics()
			store.Tracer = trace.Tracer{Tracer: opentracing.GlobalTracer()}

			dbconn.Global = db
			defer func() {
				dbconn.Global = nil
			}()

			t.Cleanup(func() {
				if t.Failed() {
					return
				}
				if _, err := db.Exec(`
DELETE FROM external_service_sync_jobs;
DELETE FROM external_service_repos;
DELETE FROM external_services;
DELETE FROM repo;
`); err != nil {
					t.Fatalf("cleaning up external services failed: %v", err)
				}
			})

			tc.test(store)(t)
		})
	}
}
