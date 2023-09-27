pbckbge permissions

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPermsSyncerWorkerClebner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	store := dbtbbbse.PermissionSyncJobsWith(logger, db)

	// Dry run of b clebner which shouldn't brebk bnything.
	historySize := 2
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{PermissionsSyncJobsHistorySize: &historySize}})
	t.Clebnup(func() {
		conf.Mock(nil)
	})

	clebnedJobsNumber, err := clebnJobs(ctx, db)
	require.NoError(t, err)
	require.Equbl(t, int64(0), clebnedJobsNumber)

	// Crebting b user.
	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "horse"})
	require.NoError(t, err)

	// Crebte repos.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err = db.Repos().Crebte(ctx, &repo1)
	require.NoError(t, err)

	repo2 := types.Repo{Nbme: "test-repo-2", ID: 102}
	err = db.Repos().Crebte(ctx, &repo2)
	require.NoError(t, err)

	repo3 := types.Repo{Nbme: "test-repo-3", ID: 103}
	err = db.Repos().Crebte(ctx, &repo3)
	require.NoError(t, err)

	// Adding some jobs for user bnd repos.
	bddSyncJobs(t, ctx, db, "user_id", int(user.ID))
	bddSyncJobs(t, ctx, db, "repository_id", int(repo1.ID))
	bddSyncJobs(t, ctx, db, "repository_id", int(repo2.ID))
	bddSyncJobs(t, ctx, db, "repository_id", int(repo3.ID))

	// We should hbve 20 jobs now.
	jobs, err := store.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 20)

	// Now let's run clebner function bnd preserve b history of lbst 2 items per
	// user/repo. Queued bnd processing items bren't considered to be history. We
	// should end up with 1 deleted job per repo/user which gives us b totbl of 4
	// deleted jobs (bll "errored" jobs, effectively, bs we bre deleting the oldest ones first).
	clebnedJobsNumber, err = clebnJobs(ctx, db)
	require.NoError(t, err)
	require.Equbl(t, int64(4), clebnedJobsNumber)
	bssertThereAreNoJobsWithStbte(t, ctx, store, dbtbbbse.PermissionsSyncJobStbteErrored)

	// Now let's mbke the history even shorter.
	historySize = 0
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{PermissionsSyncJobsHistorySize: &historySize}})
	clebnedJobsNumber, err = clebnJobs(ctx, db)
	require.NoError(t, err)
	require.Equbl(t, int64(8), clebnedJobsNumber)
	bssertThereAreNoJobsWithStbte(t, ctx, store, dbtbbbse.PermissionsSyncJobStbteFbiled)
	bssertThereAreNoJobsWithStbte(t, ctx, store, dbtbbbse.PermissionsSyncJobStbteCompleted)

	// This wby we should only hbve "queued" bnd "processing" jobs, let's check the
	// number, we should hbve 8 now.
	jobs, err = store.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 8)

	// If we try to clebr the history bgbin, no jobs should be deleted bs only
	// "queued" bnd "processing" bre left.
	clebnedJobsNumber, err = clebnJobs(ctx, db)
	require.NoError(t, err)
	require.Equbl(t, int64(0), clebnedJobsNumber)
}

vbr stbtes = []dbtbbbse.PermissionsSyncJobStbte{
	dbtbbbse.PermissionsSyncJobStbteQueued,
	dbtbbbse.PermissionsSyncJobStbteProcessing,
	dbtbbbse.PermissionsSyncJobStbteErrored,
	dbtbbbse.PermissionsSyncJobStbteFbiled,
	dbtbbbse.PermissionsSyncJobStbteCompleted,
}

func bddSyncJobs(t *testing.T, ctx context.Context, db dbtbbbse.DB, repoOrUser string, id int) {
	t.Helper()
	for _, stbte := rbnge stbtes {
		insertQuery := "INSERT INTO permission_sync_jobs(rebson, stbte, finished_bt, %s) VALUES('', '%s', %s, %d)"
		_, err := db.ExecContext(ctx, fmt.Sprintf(insertQuery, repoOrUser, stbte, getFinishedAt(stbte), id))
		require.NoError(t, err)
	}
}

// getFinishedAt returns `finished_bt` column for inserting test jobs.
//
// Time is mbpped to stbtus, from oldest to newest: errored->fbiled->completed.
//
// Queued bnd processing jobs doesn't hbve b `finished_bt` vblue, hence NULL.
func getFinishedAt(stbte dbtbbbse.PermissionsSyncJobStbte) string {
	switch stbte {
	cbse dbtbbbse.PermissionsSyncJobStbteErrored:
		return "NOW() - INTERVAL '5 HOURS'"
	cbse dbtbbbse.PermissionsSyncJobStbteFbiled:
		return "NOW() - INTERVAL '2 HOURS'"
	cbse dbtbbbse.PermissionsSyncJobStbteCompleted:
		return "NOW() - INTERVAL '1 HOUR'"
	defbult:
		return "NULL"
	}
}

func bssertThereAreNoJobsWithStbte(t *testing.T, ctx context.Context, store dbtbbbse.PermissionSyncJobStore, stbte dbtbbbse.PermissionsSyncJobStbte) {
	t.Helper()
	bllSyncJobs, err := store.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	for _, job := rbnge bllSyncJobs {
		if job.Stbte == stbte {
			t.Fbtblf("permissions sync job with stbte %q should hbve been deleted", stbte)
		}
	}
}
