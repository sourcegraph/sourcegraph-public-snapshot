package repos_test

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	for _, tc := range []struct {
		name string
		test func(database.DB) func(*testing.T)
	}{
		// {"SyncRateLimiters", testSyncRateLimiters},
		// {"EnqueueSyncJobs", testStoreEnqueueSyncJobs},
		// {"EnqueueSingleSyncJob", testStoreEnqueueSingleSyncJob},
		// {"ListExternalServiceUserIDsByRepoID", testStoreListExternalServiceUserIDsByRepoID},
		// {"ListExternalServicePrivateRepoIDsByUserID", testStoreListExternalServicePrivateRepoIDsByUserID},
		// {"Syncer/SyncWorker", testSyncWorkerPlumbing},
		// {"Syncer/Sync", testSyncerSync},
		{"Syncer/SyncWebhookWorker", testSyncWebhookWorker},
		// {"Syncer/SyncRepo", testSyncRepo},
		// {"Syncer/Run", testSyncRun},
		// {"Syncer/MultipleServices", testSyncerMultipleServices},
		// {"Syncer/OrphanedRepos", testOrphanedRepo},
		// {"Syncer/CloudDefaultExternalServicesDontSync", testCloudDefaultExternalServicesDontSync},
		// {"Syncer/DeleteExternalService", testDeleteExternalService},
		// {"Syncer/AbortSyncWhenThereIsRepoLimitError", testAbortSyncWhenThereIsRepoLimitError},
		// {"Syncer/UserAndOrgReposAreCountedCorrectly", testUserAndOrgReposAreCountedCorrectly},
		// {"Syncer/UserAddedRepos", testUserAddedRepos},
		// {"Syncer/NameConflictOnRename", testNameOnConflictOnRename},
		// {"Syncer/ConflictingSyncers", testConflictingSyncers},
		// {"Syncer/SyncRepoMaintainsOtherSources", testSyncRepoMaintainsOtherSources},
		// {"Syncer/SyncReposWithLastErrors", testSyncReposWithLastErrors},
		// {"Syncer/SyncReposWithLastErrorsHitRateLimit", testSyncReposWithLastErrorsHitsRateLimiter},
	} {
		// t.Run(tc.name, func(t *testing.T) {
		// 	store := repos.NewStore(logtest.Scoped(t), database.NewDB(dbtest.NewDB(t)))

		// 	store.SetMetrics(repos.NewStoreMetrics())
		// 	store.SetTracer(trace.Tracer{Tracer: opentracing.GlobalTracer()})

		// 	tc.test(store)(t)
		// })
		t.Run(tc.name, func(t *testing.T) {
			store := database.NewDB(dbtest.NewDB(t))
			tc.test(store)(t)
		})
	}
}
