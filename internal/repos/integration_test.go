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

	t.Parallel()

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
		{"DBStore/Syncer/SyncWorker", testSyncWorkerPlumbing},

		{"DBStore/Syncer/Batch/Sync", testSyncerBatchSync},
		{"DBStore/Syncer/Streaming/Sync", testSyncerStreamingSync},

		{"DBStore/Syncer/Batch/SyncRepo", testBatchSyncRepo},
		{"DBStore/Syncer/Streaming/SyncRepo", testStreamingSyncRepo},

		{"DBStore/Syncer/Batch/Run", testBatchSyncRun},
		{"DBStore/Syncer/Streaming/Run", testStreamingSyncRun},

		{"DBStore/Syncer/Batch/MultipleServices", testBatchSyncerMultipleServices},
		{"DBStore/Syncer/Streaming/MultipleServices", testStreamingSyncerMultipleServices},

		{"DBStore/Syncer/Batch/OrphanedRepos", testBatchOrphanedRepo},
		{"DBStore/Syncer/Streaming/OrphanedRepos", testStreamingOrphanedRepo},

		{"DBStore/Syncer/Batch/UserAddedRepos", testUserAddedRepos},
		{"DBStore/Syncer/Batch/DeleteExternalService", testDeleteExternalService},
		{"DBStore/Syncer/Batch/NameConflictDiscardOld", testNameOnConflictDiscardOld},
		{"DBStore/Syncer/Batch/NameConflictDiscardNew", testNameOnConflictDiscardNew},
		{"DBStore/Syncer/Batch/NameConflictOnRename", testNameOnConflictOnRename},
		{"DBStore/Syncer/Batch/ConflictingSyncers", testConflictingSyncers},
		{"DBStore/Syncer/Batch/SyncRepoMaintainsOtherSources", testSyncRepoMaintainsOtherSources},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := dbtest.NewDB(t, *dsn)
			dbconn.Global = db

			store := repos.NewStore(db, sql.TxOptions{Isolation: sql.LevelReadCommitted})

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			store.Log = lg
			store.Metrics = repos.NewStoreMetrics()
			store.Tracer = trace.Tracer{Tracer: opentracing.GlobalTracer()}

			t.Cleanup(func() { dbconn.Global = nil })

			tc.test(store)(t)
		})
	}
}
