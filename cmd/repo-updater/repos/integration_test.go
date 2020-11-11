package repos_test

import (
	"database/sql"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db := dbtest.NewDB(t, *dsn)

	store := repos.NewStore(db, sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	lg := log15.New()
	lg.SetHandler(log15.DiscardHandler())
	store.Log = lg

	userID := insertTestUser(t, db)

	for _, tc := range []struct {
		name string
		test func(*testing.T, *repos.Store) func(*testing.T)
	}{
		{"DBStore/SyncRateLimiters", testSyncRateLimiters},
		{"DBStore/Syncer/Sync", testSyncerSync},
		{"DBStore/Syncer/SyncWithErrors", testSyncerSyncWithErrors},
		{"DBStore/Syncer/SyncRepo", testSyncRepo},
		{"DBStore/Syncer/SyncWorker", testSyncWorkerPlumbing(db)},
		{"DBStore/Syncer/Run", testSyncRun(db)},
		{"DBStore/Syncer/MultipleServices", testSyncer(db)},
		{"DBStore/Syncer/OrphanedRepos", testOrphanedRepo(db)},
		{"DBStore/Syncer/UserAddedRepos", testUserAddedRepos(db, userID)},
		{"DBStore/Syncer/DeleteExternalService", testDeleteExternalService(db)},
		{"DBStore/Syncer/NameConflictDiscardOld", testNameOnConflictDiscardOld(db)},
		{"DBStore/Syncer/NameConflictDiscardNew", testNameOnConflictDiscardNew(db)},
		{"DBStore/Syncer/NameConflictOnRename", testNameOnConflictOnRename(db)},
		{"DBStore/Syncer/ConflictingSyncers", testConflictingSyncers(db)},
	} {
		t.Run(tc.name, func(t *testing.T) {
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

			tc.test(t, store)(t)
		})
	}
}
