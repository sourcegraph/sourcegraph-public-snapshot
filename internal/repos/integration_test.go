package repos_test

import (
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/log/logtest"

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

	t.Parallel()

	for _, tc := range []struct {
		name string
		test func(repos.Store) func(*testing.T)
	}{
		{"SyncRateLimiters", testSyncRateLimiters},
		{"EnqueueSyncJobs", testStoreEnqueueSyncJobs},
		{"EnqueueSingleSyncJob", testStoreEnqueueSingleSyncJob},
		{"EnqueueSingleWebhookBuildJob", testStoreEnqueueSingleWebhookBuildJob},
		{"ListExternalServiceUserIDsByRepoID", testStoreListExternalServiceUserIDsByRepoID},
		{"ListExternalServicePrivateRepoIDsByUserID", testStoreListExternalServicePrivateRepoIDsByUserID},
		// {"Syncer/SyncWorker", testSyncWorkerPlumbing},
		// {"Syncer/Sync", testSyncerSync},
		// {"Syncer/SyncRepo", testSyncRepo}, // PROBLEM WITHOUT foreign key
		{"Syncer/Run", testSyncRun},
		// {"Syncer/MultipleServices", testSyncerMultipleServices}, // PROBLEM with syncer.go
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := repos.NewStore(logtest.Scoped(t), database.NewDB(dbtest.NewDB(t)))

			store.SetMetrics(repos.NewStoreMetrics())
			store.SetTracer(trace.Tracer{Tracer: opentracing.GlobalTracer()})

			tc.test(store)(t)
		})
	}
}

func TestIntegration_WebhookBuilder(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	for _, tc := range []struct {
		name string
		test func(repos.Store, database.DB) func(*testing.T)
	}{
		{"WebhookBuilder", testWebhookBuilder},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := database.NewDB(dbtest.NewDB(t))
			store := repos.NewStore(logtest.Scoped(t), db)

			store.SetMetrics(repos.NewStoreMetrics())
			store.SetTracer(trace.Tracer{Tracer: opentracing.GlobalTracer()})

			tc.test(store, db)(t)
		})
	}
}
