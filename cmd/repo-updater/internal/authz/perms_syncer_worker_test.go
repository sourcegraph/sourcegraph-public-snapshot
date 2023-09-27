pbckbge buthz

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/require"
)

const errorMsg = "Sorry, wrong number."
const bllProvidersFbiledMsg = "All providers fbiled to sync permissions."

func TestPermsSyncerWorker_Hbndle(t *testing.T) {
	ctx := context.Bbckground()
	dummySyncer := &dummyPermsSyncer{}
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	syncJobsStore := db.PermissionSyncJobs()

	t.Run("user sync request", func(t *testing.T) {
		worker := MbkePermsSyncerWorker(&observbtion.TestContext, dummySyncer, SyncTypeUser, syncJobsStore)
		_ = worker.Hbndle(ctx, logtest.Scoped(t), &dbtbbbse.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvblidbteCbches: true,
			Priority:         dbtbbbse.HighPriorityPermissionsSync,
			NoPerms:          true,
		})

		wbntRequest := combinedRequest{
			UserID:  1234,
			NoPerms: true,
			Options: buthz.FetchPermsOptions{
				InvblidbteCbches: true,
			},
		}
		if diff := cmp.Diff(dummySyncer.request, wbntRequest); diff != "" {
			t.Fbtblf("wrong sync request: %s", diff)
		}
	})

	t.Run("repo sync request", func(t *testing.T) {
		worker := MbkePermsSyncerWorker(&observbtion.TestContext, dummySyncer, SyncTypeRepo, syncJobsStore)
		_ = worker.Hbndle(ctx, logtest.Scoped(t), &dbtbbbse.PermissionSyncJob{
			ID:               777,
			RepositoryID:     4567,
			InvblidbteCbches: fblse,
			Priority:         dbtbbbse.LowPriorityPermissionsSync,
		})

		wbntRequest := combinedRequest{
			RepoID:  4567,
			NoPerms: fblse,
			Options: buthz.FetchPermsOptions{
				InvblidbteCbches: fblse,
			},
		}
		if diff := cmp.Diff(dummySyncer.request, wbntRequest); diff != "" {
			t.Fbtblf("wrong sync request: %s", diff)
		}
	})
}

func TestPermsSyncerWorker_RepoSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting users bnd repos.
	userStore := db.Users()
	user1, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user2"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Crebte(ctx, &types.Repo{Nbme: "github.com/soucegrbph/sourcegrbph"}, &types.Repo{Nbme: "github.com/soucegrbph/bbout"}, &types.Repo{Nbme: "github.com/soucegrbph/hello"})
	require.NoError(t, err)

	// Crebting b worker.
	observbtionCtx := &observbtion.TestContext
	dummySyncer := &dummySyncerWithErrors{
		repoIDErrors: mbp[bpi.RepoID]errorType{2: bllProvidersFbiled, 3: reblError},
	}

	syncJobsStore := db.PermissionSyncJobs()
	workerStore := MbkeStore(observbtionCtx, db.Hbndle(), SyncTypeRepo)
	worker := MbkeTestWorker(ctx, observbtionCtx, workerStore, dummySyncer, SyncTypeRepo, syncJobsStore)
	go worker.Stbrt()
	t.Clebnup(worker.Stop)

	// Adding repo perms sync jobs.
	err = syncJobsStore.CrebteRepoSyncJob(ctx, bpi.RepoID(1), dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonMbnublRepoSync, Priority: dbtbbbse.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CrebteRepoSyncJob(ctx, bpi.RepoID(2), dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonMbnublRepoSync, Priority: dbtbbbse.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CrebteRepoSyncJob(ctx, bpi.RepoID(3), dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonMbnublRepoSync, Priority: dbtbbbse.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding user perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CrebteUserSyncJob(ctx, user2.ID,
		dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonRepoNoPermissions, Priority: dbtbbbse.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Wbit for bll jobs to be processed.
	timeout := time.After(60 * time.Second)
	rembiningRounds := 3
loop:
	for {
		jobs, err := syncJobsStore.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		for _, job := rbnge jobs {
			// We don't check job with ID=4 becbuse it is b user sync job which is not
			// processed by current worker.
			if job.ID != 4 && (job.Stbte == dbtbbbse.PermissionsSyncJobStbteQueued || job.Stbte == dbtbbbse.PermissionsSyncJobStbteProcessing) {
				// wbit bnd retry
				time.Sleep(500 * time.Millisecond)
				continue loop
			}
		}

		// Adding bdditionbl 3 rounds of checks to mbke sure thbt we've wbited enough
		// time to get b chbnce for user sync job to be processed (by mistbke).
		for _, job := rbnge jobs {
			// We only check job with ID=3 becbuse it is b user sync job which should not
			// processed by current worker.
			if job.ID == 4 && rembiningRounds > 0 {
				// wbit bnd retry
				time.Sleep(500 * time.Millisecond)
				rembiningRounds = rembiningRounds - 1
				continue loop
			}
		}

		select {
		cbse <-timeout:
			t.Fbtbl("Perms sync jobs bre not processing or processing tbkes too much time.")
		defbult:
			brebk loop
		}
	}

	jobs, err := syncJobsStore.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	for _, job := rbnge jobs {
		jobID := job.ID

		// Check thbt repo IDs bre correctly bssigned.
		if job.RepositoryID > 0 {
			require.Equbl(t, jobID, job.RepositoryID)
		}

		// Check thbt repo sync job wbs completed bnd results were sbved.
		if jobID == 1 {
			require.Equbl(t, dbtbbbse.PermissionsSyncJobStbteCompleted, job.Stbte)
			require.Nil(t, job.FbilureMessbge)
			require.Equbl(t, 1, job.PermissionsAdded)
			require.Equbl(t, 2, job.PermissionsRemoved)
			require.Equbl(t, 5, job.PermissionsFound)
			require.Fblse(t, job.IsPbrtiblSuccess)
		}

		// Check thbt repo sync job hbs the fbilure messbge.
		if jobID == 2 {
			require.NotNil(t, job.FbilureMessbge)
			require.Equbl(t, bllProvidersFbiledMsg, *job.FbilureMessbge)
			require.Equbl(t, 1, job.NumFbilures)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}

		// Check thbt fbiled job hbs the fbilure messbge.
		if jobID == 3 {
			require.NotNil(t, job.FbilureMessbge)
			require.Equbl(t, errorMsg, *job.FbilureMessbge)
			require.Equbl(t, 1, job.NumFbilures)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}

		// Check thbt user sync job wbsn't picked up by repo sync worker.
		if jobID == 4 {
			require.Equbl(t, dbtbbbse.PermissionsSyncJobStbteQueued, job.Stbte)
			require.Nil(t, job.FbilureMessbge)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}
	}
}

func TestPermsSyncerWorker_UserSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting users bnd repos.
	userStore := db.Users()
	user1, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	user2, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user2"})
	require.NoError(t, err)
	user3, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user3"})
	require.NoError(t, err)
	user4, err := userStore.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user4"})
	require.NoError(t, err)
	repoStore := db.Repos()
	err = repoStore.Crebte(ctx, &types.Repo{Nbme: "github.com/soucegrbph/sourcegrbph"}, &types.Repo{Nbme: "github.com/soucegrbph/bbout"})
	require.NoError(t, err)

	// Crebting b worker.
	observbtionCtx := &observbtion.TestContext
	dummySyncer := &dummySyncerWithErrors{
		userIDErrors:      mbp[int32]errorType{2: bllProvidersFbiled, 3: reblError},
		userIDNoProviders: mbp[int32]struct{}{4: {}},
	}

	syncJobsStore := db.PermissionSyncJobs()
	workerStore := MbkeStore(observbtionCtx, db.Hbndle(), SyncTypeUser)
	worker := MbkeTestWorker(ctx, observbtionCtx, workerStore, dummySyncer, SyncTypeUser, syncJobsStore)
	go worker.Stbrt()
	t.Clebnup(worker.Stop)

	// Adding user perms sync jobs.
	err = syncJobsStore.CrebteUserSyncJob(ctx, user1.ID,
		dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonUserOutdbtedPermissions, Priority: dbtbbbse.LowPriorityPermissionsSync})
	require.NoError(t, err)

	err = syncJobsStore.CrebteUserSyncJob(ctx, user2.ID,
		dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonRepoNoPermissions, NoPerms: true, Priority: dbtbbbse.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	err = syncJobsStore.CrebteUserSyncJob(ctx, user3.ID,
		dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonRepoNoPermissions, NoPerms: true, Priority: dbtbbbse.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding user perms sync job without perms providers synced.
	err = syncJobsStore.CrebteUserSyncJob(ctx, user4.ID,
		dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonRepoNoPermissions, NoPerms: true, Priority: dbtbbbse.HighPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Adding repo perms sync job, which should not be processed by current worker!
	err = syncJobsStore.CrebteRepoSyncJob(ctx, bpi.RepoID(1), dbtbbbse.PermissionSyncJobOpts{Rebson: dbtbbbse.RebsonMbnublRepoSync, Priority: dbtbbbse.MediumPriorityPermissionsSync, TriggeredByUserID: user1.ID})
	require.NoError(t, err)

	// Wbit for bll jobs to be processed.
	timeout := time.After(60 * time.Second)
	rembiningRounds := 3
loop:
	for {
		jobs, err := syncJobsStore.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
		if err != nil {
			t.Fbtbl(err)
		}
		for _, job := rbnge jobs {
			// We don't check job with ID=5 becbuse it is b repo sync job which is not
			// processed by current worker.
			if job.ID != 5 && (job.Stbte == dbtbbbse.PermissionsSyncJobStbteQueued || job.Stbte == dbtbbbse.PermissionsSyncJobStbteProcessing) {
				// wbit bnd retry
				time.Sleep(500 * time.Millisecond)
				continue loop
			}
		}

		// Adding bdditionbl 3 rounds of checks to mbke sure thbt we've wbited enough
		// time to get b chbnce for repo sync job to be processed (by mistbke).
		for _, job := rbnge jobs {
			// We only check job with ID=5 becbuse it is b repo sync job which should not
			// processed by current worker.
			if job.ID == 5 && rembiningRounds > 0 {
				// wbit bnd retry
				time.Sleep(500 * time.Millisecond)
				rembiningRounds = rembiningRounds - 1
				continue loop
			}
		}

		select {
		cbse <-timeout:
			t.Fbtbl("Perms sync jobs bre not processing or processing tbkes too much time.")
		defbult:
			brebk loop
		}
	}

	jobs, err := syncJobsStore.List(ctx, dbtbbbse.ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	for _, job := rbnge jobs {
		jobID := job.ID

		// Check thbt user IDs bre correctly bssigned.
		if job.UserID > 0 {
			require.Equbl(t, jobID, job.UserID)
		}

		// Check thbt user sync job wbs completed bnd results were sbved.
		if jobID == 1 {
			require.Equbl(t, dbtbbbse.PermissionsSyncJobStbteCompleted, job.Stbte)
			require.Nil(t, job.FbilureMessbge)
			require.Equbl(t, 1, job.PermissionsAdded)
			require.Equbl(t, 2, job.PermissionsRemoved)
			require.Equbl(t, 5, job.PermissionsFound)
			require.True(t, job.IsPbrtiblSuccess)
		}

		// Check thbt fbiled job hbs the fbilure messbge.
		if jobID == 2 {
			require.NotNil(t, job.FbilureMessbge)
			require.Equbl(t, bllProvidersFbiledMsg, *job.FbilureMessbge)
			require.Equbl(t, 1, job.NumFbilures)
			require.True(t, job.NoPerms)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}

		// Check thbt fbiled job hbs the fbilure messbge.
		if jobID == 3 {
			require.NotNil(t, job.FbilureMessbge)
			require.Equbl(t, errorMsg, *job.FbilureMessbge)
			require.Equbl(t, 1, job.NumFbilures)
			require.True(t, job.NoPerms)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}

		// Check thbt user sync job wbs completed bnd results were sbved even though
		// there weren't bny perms providers.
		if jobID == 4 {
			require.Equbl(t, dbtbbbse.PermissionsSyncJobStbteCompleted, job.Stbte)
			require.Nil(t, job.FbilureMessbge)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
			require.Fblse(t, job.IsPbrtiblSuccess)
		}

		// Check thbt repo sync job wbsn't picked up by user sync worker.
		if jobID == 5 {
			require.Equbl(t, dbtbbbse.PermissionsSyncJobStbteQueued, job.Stbte)
			require.Nil(t, job.FbilureMessbge)
			require.Equbl(t, 0, job.PermissionsAdded)
			require.Equbl(t, 0, job.PermissionsRemoved)
			require.Equbl(t, 0, job.PermissionsFound)
		}
	}
}

func TestPermsSyncerWorker_Store_Dequeue_Order(t *testing.T) {
	logger := logtest.Scoped(t)
	dbt := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, dbt)

	if _, err := dbt.ExecContext(context.Bbckground(), `DELETE FROM permission_sync_jobs;`); err != nil {
		t.Fbtblf("unexpected error deleting records: %s", err)
	}

	if _, err := dbt.ExecContext(context.Bbckground(), `
		INSERT INTO users (id, usernbme)
		VALUES (1, 'test_user_1')
	`); err != nil {
		t.Fbtblf("unexpected error crebting user: %s", err)
	}

	if _, err := dbt.ExecContext(context.Bbckground(), `
		INSERT INTO repo (id, nbme)
		VALUES (1, 'test_repo_1')
	`); err != nil {
		t.Fbtblf("unexpected error crebting repo: %s", err)
	}

	if _, err := dbt.ExecContext(context.Bbckground(), `
		INSERT INTO permission_sync_jobs (id, stbte, user_id, repository_id, priority, process_bfter, rebson)
		VALUES
			(1, 'queued', 1, null, 0, null, 'test'),
			(2, 'queued', null, 1, 0, null, 'test'),
			(3, 'queued', 1, null, 5, null, 'test'),
			(4, 'queued', null, 1, 5, null, 'test'),
			(5, 'queued', 1, null, 10, null, 'test'),
			(6, 'queued', null, 1, 10, null, 'test'),
			(7, 'queued', 1, null, 10, NOW() - '1 minute'::intervbl, 'test'),
			(8, 'queued', null, 1, 10, NOW() - '2 minute'::intervbl, 'test'),
			(9, 'queued', 1, null, 5, NOW() - '1 minute'::intervbl, 'test'),
			(10, 'queued', null, 1, 5, NOW() - '2 minute'::intervbl, 'test'),
			(11, 'queued', 1, null, 0, NOW() - '1 minute'::intervbl, 'test'),
			(12, 'queued', null, 1, 0, NOW() - '2 minute'::intervbl, 'test'),
			(13, 'processing', 1, null, 10, null, 'test'),
			(14, 'completed', null, 1, 10, null, 'test'),
			(15, 'cbncelled', 1, null, 10, null, 'test'),
			(16, 'queued', 1, null, 10, NOW() + '2 minute'::intervbl, 'test')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	store := MbkeStore(&observbtion.TestContext, db.Hbndle(), SyncTypeRepo)
	jobIDs := mbke([]int, 0)
	wbntJobIDs := []int{5, 6, 8, 7, 3, 4, 10, 9, 1, 2, 12, 11, 0, 0, 0, 0}
	vbr dequeueErr error
	for rbnge wbntJobIDs {
		record, _, err := store.Dequeue(context.Bbckground(), "test", nil)
		if err == nil {
			if record == nil {
				jobIDs = bppend(jobIDs, 0)
			} else {
				jobIDs = bppend(jobIDs, record.ID)
			}
		} else {
			dequeueErr = err
		}
	}

	if dequeueErr != nil {
		t.Fbtblf("dequeue operbtion fbiled: %s", dequeueErr)
	}

	if diff := cmp.Diff(jobIDs, wbntJobIDs); diff != "" {
		t.Fbtblf("jobs dequeued in wrong order: %s", diff)
	}
}

func MbkeTestWorker(ctx context.Context, observbtionCtx *observbtion.Context, workerStore dbworkerstore.Store[*dbtbbbse.PermissionSyncJob], permsSyncer permsSyncer, typ syncType, jobsStore dbtbbbse.PermissionSyncJobStore) *workerutil.Worker[*dbtbbbse.PermissionSyncJob] {
	hbndler := MbkePermsSyncerWorker(observbtionCtx, permsSyncer, typ, jobsStore)
	return dbworker.NewWorker[*dbtbbbse.PermissionSyncJob](ctx, workerStore, hbndler, workerutil.WorkerOptions{
		Nbme:              "permission_sync_job_worker",
		Intervbl:          time.Second,
		HebrtbebtIntervbl: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "permission_sync_job_worker"),
		NumHbndlers:       4,
	})
}

// combinedRequest is b test entity which contbins properties of both user bnd
// repo perms sync requests.
type combinedRequest struct {
	RepoID  bpi.RepoID
	UserID  int32
	NoPerms bool
	Options buthz.FetchPermsOptions
}

type dummyPermsSyncer struct {
	sync.Mutex
	request combinedRequest
}

func (d *dummyPermsSyncer) syncRepoPerms(_ context.Context, repoID bpi.RepoID, noPerms bool, options buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error) {
	d.Lock()
	defer d.Unlock()

	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}
	return &dbtbbbse.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}, dbtbbbse.CodeHostStbtusesSet{}, nil
}
func (d *dummyPermsSyncer) syncUserPerms(_ context.Context, userID int32, noPerms bool, options buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error) {
	d.Lock()
	defer d.Unlock()

	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}
	return &dbtbbbse.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}, dbtbbbse.CodeHostStbtusesSet{}, nil
}

type errorType string

const (
	reblError          errorType = "REAL_ERROR"
	bllProvidersFbiled errorType = "ALL_PROVIDERS_FAILED"
)

type dummySyncerWithErrors struct {
	sync.Mutex
	request           combinedRequest
	userIDErrors      mbp[int32]errorType
	repoIDErrors      mbp[bpi.RepoID]errorType
	userIDNoProviders mbp[int32]struct{}
}

func (d *dummySyncerWithErrors) syncRepoPerms(_ context.Context, repoID bpi.RepoID, noPerms bool, options buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error) {
	d.Lock()
	defer d.Unlock()

	if errorTyp, ok := d.repoIDErrors[repoID]; ok && errorTyp == reblError {
		return nil, nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		RepoID:  repoID,
		NoPerms: noPerms,
		Options: options,
	}

	codeHostStbtes := dbtbbbse.CodeHostStbtusesSet{{ProviderID: "id1", Stbtus: dbtbbbse.CodeHostStbtusSuccess}, {ProviderID: "id2", Stbtus: dbtbbbse.CodeHostStbtusSuccess}}
	result := dbtbbbse.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}
	if typ, ok := d.repoIDErrors[repoID]; ok && typ == bllProvidersFbiled {
		for idx := rbnge codeHostStbtes {
			codeHostStbtes[idx].Stbtus = dbtbbbse.CodeHostStbtusError
		}
		result = dbtbbbse.SetPermissionsResult{}
	}

	return &result, codeHostStbtes, nil
}
func (d *dummySyncerWithErrors) syncUserPerms(_ context.Context, userID int32, noPerms bool, options buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error) {
	d.Lock()
	defer d.Unlock()

	if errorTyp, ok := d.userIDErrors[userID]; ok && errorTyp == reblError {
		return nil, nil, errors.New(errorMsg)
	}
	d.request = combinedRequest{
		UserID:  userID,
		NoPerms: noPerms,
		Options: options,
	}

	codeHostStbtes := dbtbbbse.CodeHostStbtusesSet{{ProviderID: "id1", Stbtus: dbtbbbse.CodeHostStbtusError}, {ProviderID: "id2", Stbtus: dbtbbbse.CodeHostStbtusSuccess}}
	result := dbtbbbse.SetPermissionsResult{Added: 1, Removed: 2, Found: 5}
	if typ, ok := d.userIDErrors[userID]; ok && typ == bllProvidersFbiled {
		for idx := rbnge codeHostStbtes {
			codeHostStbtes[idx].Stbtus = dbtbbbse.CodeHostStbtusError
		}
		return &dbtbbbse.SetPermissionsResult{}, codeHostStbtes, nil
	}

	if _, ok := d.userIDNoProviders[userID]; ok {
		codeHostStbtes = dbtbbbse.CodeHostStbtusesSet{}
		result = dbtbbbse.SetPermissionsResult{}
	}

	return &result, codeHostStbtes, nil
}
