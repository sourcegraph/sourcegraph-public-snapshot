package repos_test

import (
	"testing"

	"github.com/sourcegraph/log/logtest"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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

	logger := logtest.Scoped(t)
	t.Parallel()

	for _, tc := range []struct {
		name string
		test func(repos.Store) func(*testing.T)
	}{
		{"SyncRateLimiters", testSyncRateLimiters},
		{"EnqueueSyncJobs", testStoreEnqueueSyncJobs},
		{"EnqueueSingleSyncJob", testStoreEnqueueSingleSyncJob},
		{"ListExternalServiceUserIDsByRepoID", testStoreListExternalServiceUserIDsByRepoID},
		{"ListExternalServicePrivateRepoIDsByUserID", testStoreListExternalServicePrivateRepoIDsByUserID},
		{"Syncer/SyncWorker", testSyncWorkerPlumbing},
		{"Syncer/Sync", testSyncerSync},
		{"Syncer/SyncRepo", testSyncRepo},
		{"Syncer/Run", testSyncRun},
		{"Syncer/MultipleServices", testSyncerMultipleServices},
		{"Syncer/OrphanedRepos", testOrphanedRepo},
		{"Syncer/CloudDefaultExternalServicesDontSync", testCloudDefaultExternalServicesDontSync},
		{"Syncer/DeleteExternalService", testDeleteExternalService},
		{"Syncer/AbortSyncWhenThereIsRepoLimitError", testAbortSyncWhenThereIsRepoLimitError},
		{"Syncer/UserAndOrgReposAreCountedCorrectly", testUserAndOrgReposAreCountedCorrectly},
		{"Syncer/UserAddedRepos", testUserAddedRepos},
		{"Syncer/NameConflictOnRename", testNameOnConflictOnRename},
		{"Syncer/ConflictingSyncers", testConflictingSyncers},
		{"Syncer/SyncRepoMaintainsOtherSources", testSyncRepoMaintainsOtherSources},
		{"Syncer/SyncReposWithLastErrors", testSyncReposWithLastErrors},
		{"Syncer/SyncReposWithLastErrorsHitRateLimit", testSyncReposWithLastErrorsHitsRateLimiter},
		{"WebhookWorker/EnqueueWebhookBuildJob", testEnqueueWebhookBuildJob},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := repos.NewStore(logtest.Scoped(t), database.NewDB(logger, dbtest.NewDB(logger, t)))

			store.SetMetrics(repos.NewStoreMetrics())
			store.SetTracer(trace.Tracer{TracerProvider: otel.GetTracerProvider()})

			tc.test(store)(t)
		})
	}
}
