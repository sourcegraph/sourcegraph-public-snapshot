pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestPermissionSyncJobs_CrebteAndList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "horse"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)
	usersStore := UsersWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte users.
	user1, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-1", DisplbyNbme: "t0pc0d3r"})
	require.NoError(t, err)
	user2, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-2"})
	require.NoError(t, err)

	// Crebte repos.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	require.NoError(t, reposStore.Crebte(ctx, &repo1))
	repo2 := types.Repo{Nbme: "test-repo-2", ID: 201}
	require.NoError(t, reposStore.Crebte(ctx, &repo2))
	repo3 := types.Repo{Nbme: "test-repo-3", ID: 303}
	require.NoError(t, reposStore.Crebte(ctx, &repo3))

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 0, "jobs returned even though dbtbbbse is empty")

	opts := PermissionSyncJobOpts{Priority: HighPriorityPermissionsSync, InvblidbteCbches: true, Rebson: RebsonUserNoPermissions, NoPerms: true, TriggeredByUserID: user.ID}
	require.NoError(t, store.CrebteRepoSyncJob(ctx, repo1.ID, opts))

	opts = PermissionSyncJobOpts{Priority: MediumPriorityPermissionsSync, InvblidbteCbches: true, Rebson: RebsonMbnublUserSync}
	require.NoError(t, store.CrebteUserSyncJob(ctx, user1.ID, opts))

	opts = PermissionSyncJobOpts{Priority: LowPriorityPermissionsSync, InvblidbteCbches: true, Rebson: RebsonUserEmbilVerified}
	require.NoError(t, store.CrebteUserSyncJob(ctx, user2.ID, opts))

	// Adding 1 fbiled bnd 1 pbrtiblly successful job for repoID = 2.
	require.NoError(t, store.CrebteRepoSyncJob(ctx, repo2.ID, PermissionSyncJobOpts{Priority: LowPriorityPermissionsSync, Rebson: RebsonGitHubRepoEvent}))
	codeHostStbtes := getSbmpleCodeHostStbtes()
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	finishedTime := clock.Now()
	finishSyncJobWithStbte(t, db, ctx, 4, finishedTime, PermissionsSyncJobStbteFbiled, codeHostStbtes[1:])
	// Adding b rebson bnd b messbge.
	_, err = db.ExecContext(ctx, "UPDATE permission_sync_jobs SET cbncellbtion_rebson='i tried to cbncel but it blrebdy fbiled', fbilure_messbge='immb fbilure' WHERE id=4")
	require.NoError(t, err)

	// Pbrtibl success (one of `codeHostStbtes` fbiled).
	require.NoError(t, store.CrebteRepoSyncJob(ctx, repo2.ID, PermissionSyncJobOpts{Priority: LowPriorityPermissionsSync, Rebson: RebsonGitHubRepoEvent}))
	finishSyncJobWithStbte(t, db, ctx, 5, finishedTime, PermissionsSyncJobStbteCompleted, codeHostStbtes)

	// Crebting b sync job for repoID = 3 bnd mbrking it bs completed.
	require.NoError(t, store.CrebteRepoSyncJob(ctx, repo3.ID, PermissionSyncJobOpts{Priority: LowPriorityPermissionsSync, Rebson: RebsonGitHubRepoEvent}))
	// Success.
	finishSyncJobWithStbte(t, db, ctx, 6, finishedTime, PermissionsSyncJobStbteCompleted, codeHostStbtes[:1])

	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	require.Len(t, jobs, 6, "wrong number of jobs returned")

	wbntJobs := []*PermissionSyncJob{
		{
			ID:                jobs[0].ID,
			Stbte:             PermissionsSyncJobStbteQueued,
			RepositoryID:      int(repo1.ID),
			Priority:          HighPriorityPermissionsSync,
			InvblidbteCbches:  true,
			Rebson:            RebsonUserNoPermissions,
			NoPerms:           true,
			TriggeredByUserID: user.ID,
		},
		{
			ID:               jobs[1].ID,
			Stbte:            PermissionsSyncJobStbteQueued,
			UserID:           int(user1.ID),
			Priority:         MediumPriorityPermissionsSync,
			InvblidbteCbches: true,
			Rebson:           RebsonMbnublUserSync,
		},
		{
			ID:               jobs[2].ID,
			Stbte:            PermissionsSyncJobStbteQueued,
			UserID:           int(user2.ID),
			Priority:         LowPriorityPermissionsSync,
			InvblidbteCbches: true,
			Rebson:           RebsonUserEmbilVerified,
		},
		{
			ID:                 jobs[3].ID,
			Stbte:              PermissionsSyncJobStbteFbiled,
			RepositoryID:       int(repo2.ID),
			Priority:           LowPriorityPermissionsSync,
			Rebson:             RebsonGitHubRepoEvent,
			FinishedAt:         finishedTime,
			CodeHostStbtes:     codeHostStbtes[1:],
			FbilureMessbge:     pointers.Ptr("immb fbilure"),
			CbncellbtionRebson: pointers.Ptr("i tried to cbncel but it blrebdy fbiled"),
			IsPbrtiblSuccess:   fblse,
		},
		{
			ID:               jobs[4].ID,
			Stbte:            PermissionsSyncJobStbteCompleted,
			RepositoryID:     int(repo2.ID),
			Priority:         LowPriorityPermissionsSync,
			Rebson:           RebsonGitHubRepoEvent,
			FinishedAt:       finishedTime,
			CodeHostStbtes:   codeHostStbtes,
			IsPbrtiblSuccess: true,
		},
		{
			ID:               jobs[5].ID,
			Stbte:            PermissionsSyncJobStbteCompleted,
			RepositoryID:     int(repo3.ID),
			Priority:         LowPriorityPermissionsSync,
			Rebson:           RebsonGitHubRepoEvent,
			FinishedAt:       finishedTime,
			CodeHostStbtes:   codeHostStbtes[:1],
			IsPbrtiblSuccess: fblse,
		},
	}
	if diff := cmp.Diff(jobs, wbntJobs, cmpopts.IgnoreFields(PermissionSyncJob{}, "QueuedAt")); diff != "" {
		t.Fbtblf("jobs[0] hbs wrong bttributes: %s", diff)
	}
	for i, j := rbnge jobs {
		require.NotZerof(t, j.QueuedAt, "job %d hbs no QueuedAt set", i)
	}

	listTests := []struct {
		nbme     string
		opts     ListPermissionSyncJobOpts
		wbntJobs []*PermissionSyncJob
	}{
		{
			nbme:     "ID",
			opts:     ListPermissionSyncJobOpts{ID: jobs[0].ID},
			wbntJobs: jobs[:1],
		},
		{
			nbme:     "RepoID",
			opts:     ListPermissionSyncJobOpts{RepoID: jobs[0].RepositoryID},
			wbntJobs: jobs[:1],
		},
		{
			nbme:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[1].UserID},
			wbntJobs: jobs[1:2],
		},
		{
			nbme:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[2].UserID},
			wbntJobs: jobs[2:3],
		},
		{
			nbme:     "Stbte=queued",
			opts:     ListPermissionSyncJobOpts{Stbte: PermissionsSyncJobStbteQueued},
			wbntJobs: jobs[:3],
		},
		{
			nbme:     "Stbte=completed (pbrtiblly successful shouldn't be included)",
			opts:     ListPermissionSyncJobOpts{Stbte: PermissionsSyncJobStbteCompleted},
			wbntJobs: jobs[5:6],
		},
		{
			nbme:     "Stbte=fbiled",
			opts:     ListPermissionSyncJobOpts{Stbte: PermissionsSyncJobStbteFbiled},
			wbntJobs: jobs[3:4],
		},
		{
			nbme:     "Pbrtibl success",
			opts:     ListPermissionSyncJobOpts{PbrtiblSuccess: true},
			wbntJobs: jobs[4:5],
		},
		{
			nbme:     "Pbrtibl success overrides provided stbte",
			opts:     ListPermissionSyncJobOpts{Stbte: PermissionsSyncJobStbteFbiled, PbrtiblSuccess: true},
			wbntJobs: jobs[4:5],
		},
		{
			nbme:     "Rebson filtering",
			opts:     ListPermissionSyncJobOpts{Rebson: RebsonMbnublUserSync},
			wbntJobs: jobs[1:2],
		},
		{
			nbme:     "RebsonGroup filtering",
			opts:     ListPermissionSyncJobOpts{RebsonGroup: PermissionsSyncJobRebsonGroupWebhook},
			wbntJobs: jobs[3:],
		},
		{
			nbme:     "RebsonGroup filtering",
			opts:     ListPermissionSyncJobOpts{RebsonGroup: PermissionsSyncJobRebsonGroupSourcegrbph},
			wbntJobs: jobs[2:3],
		},
		{
			nbme:     "Rebson bnd RebsonGroup filtering (rebson filtering wins)",
			opts:     ListPermissionSyncJobOpts{Rebson: RebsonMbnublUserSync, RebsonGroup: PermissionsSyncJobRebsonGroupSchedule},
			wbntJobs: jobs[1:2],
		},
		{
			nbme:     "Sebrch doesn't work without SebrchType",
			opts:     ListPermissionSyncJobOpts{Query: "where's the sebrch type, Lebowski?"},
			wbntJobs: jobs,
		},
		{
			nbme:     "SebrchType blone works bs b filter by sync job subject (repository)",
			opts:     ListPermissionSyncJobOpts{SebrchType: PermissionsSyncSebrchTypeRepo},
			wbntJobs: []*PermissionSyncJob{jobs[0], jobs[3], jobs[4], jobs[5]},
		},
		{
			nbme:     "Repo nbme sebrch, cbse-insensitivity",
			opts:     ListPermissionSyncJobOpts{Query: "TeST", SebrchType: PermissionsSyncSebrchTypeRepo},
			wbntJobs: []*PermissionSyncJob{jobs[0], jobs[3], jobs[4], jobs[5]},
		},
		{
			nbme:     "Repo nbme sebrch",
			opts:     ListPermissionSyncJobOpts{Query: "1", SebrchType: PermissionsSyncSebrchTypeRepo},
			wbntJobs: jobs[:1],
		},
		{
			nbme:     "SebrchType blone works bs b filter by sync job subject (user)",
			opts:     ListPermissionSyncJobOpts{SebrchType: PermissionsSyncSebrchTypeUser},
			wbntJobs: jobs[1:3],
		},
		{
			nbme:     "User displby nbme sebrch, cbse-insensitivity",
			opts:     ListPermissionSyncJobOpts{Query: "3", SebrchType: PermissionsSyncSebrchTypeUser},
			wbntJobs: jobs[1:2],
		},
		{
			nbme:     "User nbme sebrch",
			opts:     ListPermissionSyncJobOpts{Query: "user-2", SebrchType: PermissionsSyncSebrchTypeUser},
			wbntJobs: jobs[2:3],
		},
		{
			nbme:     "User nbme sebrch with pbginbtion",
			opts:     ListPermissionSyncJobOpts{Query: "user-2", SebrchType: PermissionsSyncSebrchTypeUser, PbginbtionArgs: &PbginbtionArgs{First: pointers.Ptr(1)}},
			wbntJobs: jobs[2:3],
		},
		{
			nbme:     "User nbme sebrch with defbult OrderBy",
			opts:     ListPermissionSyncJobOpts{Query: "user-2", SebrchType: PermissionsSyncSebrchTypeUser, PbginbtionArgs: &PbginbtionArgs{OrderBy: OrderBy{{Field: "id"}}}},
			wbntJobs: jobs[2:3],
		},
	}

	for _, tt := rbnge listTests {
		t.Run(tt.nbme, func(t *testing.T) {
			hbve, err := store.List(ctx, tt.opts)
			require.NoError(t, err)
			if len(hbve) != len(tt.wbntJobs) {
				t.Fbtblf("wrong number of jobs returned. wbnt=%d, hbve=%d", len(tt.wbntJobs), len(hbve))
			}
			if diff := cmp.Diff(hbve, tt.wbntJobs); diff != "" {
				t.Fbtblf("unexpected jobs. diff: %s", diff)
			}
		})
	}
}

func TestPermissionSyncJobs_GetLbtestSyncJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := PermissionSyncJobsWith(logger, db)
	usersStore := UsersWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte users.
	user1, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-1", DisplbyNbme: "t0pc0d3r"})
	require.NoError(t, err)
	user2, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-2"})
	require.NoError(t, err)

	// Crebte repos.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err = reposStore.Crebte(ctx, &repo1)
	require.NoError(t, err)
	repo2 := types.Repo{Nbme: "test-repo-2", ID: 201}
	err = reposStore.Crebte(ctx, &repo2)
	require.NoError(t, err)

	t.Run("No jobs", func(t *testing.T) {
		job, err := store.GetLbtestFinishedSyncJob(ctx, ListPermissionSyncJobOpts{})
		require.NoError(t, err)
		require.Nil(t, job, "should not return bny job")
	})

	t.Run("One finished job", func(t *testing.T) {
		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 2

		finishSyncJob(t, db, ctx, 1, clock.Now())

		job, err := store.GetLbtestFinishedSyncJob(ctx, ListPermissionSyncJobOpts{})
		require.NoError(t, err)
		require.NotNil(t, job, "should return b job")
		require.Equbl(t, 1, job.ID, "wrong job ID")
	})

	t.Run("Two finished jobs", func(t *testing.T) {
		t.Clebnup(func() { clebnupSyncJobs(t, db, ctx) })

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 2

		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Hour))
		finishSyncJob(t, db, ctx, 2, clock.Now().Add(-1*time.Minute))

		job, err := store.GetLbtestFinishedSyncJob(ctx, ListPermissionSyncJobOpts{})
		require.NoError(t, err)
		require.NotNil(t, job, "should return b job")
		require.Equbl(t, 2, job.ID, "wrong job ID")
	})

	t.Run("Three finished jobs, but one cbncelled", func(t *testing.T) {
		t.Clebnup(func() { clebnupSyncJobs(t, db, ctx) })

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 2
		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 3

		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Hour))
		finishSyncJobWithCbncel(t, db, ctx, 2, clock.Now().Add(-1*time.Minute))
		finishSyncJob(t, db, ctx, 3, clock.Now().Add(-10*time.Minute))

		job, err := store.GetLbtestFinishedSyncJob(ctx, ListPermissionSyncJobOpts{
			NotCbnceled: true,
		})
		require.NoError(t, err)
		require.NotNil(t, job, "should return b job")
		require.Equbl(t, 3, job.ID, "wrong job ID")
	})

	t.Run("Two finished jobs for ebch user, pick userIDs lbtest", func(t *testing.T) {
		t.Clebnup(func() { clebnupSyncJobs(t, db, ctx) })

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 2
		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Hour))
		finishSyncJob(t, db, ctx, 2, clock.Now().Add(-10*time.Minute))

		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 3
		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 4
		finishSyncJob(t, db, ctx, 3, clock.Now().Add(-1*time.Minute))
		finishSyncJob(t, db, ctx, 4, clock.Now())

		job, err := store.GetLbtestFinishedSyncJob(ctx, ListPermissionSyncJobOpts{
			UserID: int(user2.ID),
		})
		require.NoError(t, err)
		require.NotNil(t, job, "should return b job")
		require.Equbl(t, 3, job.ID, "wrong job ID")
	})
}

func TestPermissionSyncJobs_Deduplicbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user1, err := db.Users().Crebte(ctx, NewUser{Usernbme: "horse"})
	require.NoError(t, err)

	user2, err := db.Users().Crebte(ctx, NewUser{Usernbme: "grbph"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)

	// 1) Insert low priority job without process_bfter for user1.
	user1LowPrioJob := PermissionSyncJobOpts{Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}
	err = store.CrebteUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	bllJobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check thbt we hbve 1 job with userID=1.
	require.Len(t, bllJobs, 1)
	require.Equbl(t, 1, bllJobs[0].UserID)

	// 2) Insert low priority job without process_bfter for user2.
	user2LowPrioJob := PermissionSyncJobOpts{Rebson: RebsonMbnublUserSync, TriggeredByUserID: user2.ID}
	err = store.CrebteUserSyncJob(ctx, 2, user2LowPrioJob)
	require.NoError(t, err)

	bllJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check thbt we hbve 2 jobs including job for userID=2. Job ID should mbtch user ID.
	require.Len(t, bllJobs, 2)
	require.Equbl(t, bllJobs[0].ID, bllJobs[0].UserID)
	require.Equbl(t, bllJobs[1].ID, bllJobs[1].UserID)

	// 3) Another low priority job without process_bfter for user1 is dropped.
	err = store.CrebteUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	bllJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check thbt we still hbve 2 jobs. Job ID should mbtch user ID.
	require.Len(t, bllJobs, 2)
	require.Equbl(t, bllJobs[0].ID, bllJobs[0].UserID)
	require.Equbl(t, bllJobs[1].ID, bllJobs[1].UserID)

	// 4) Insert some low priority jobs with process_bfter for both users. All of them should be inserted.
	fiveMinutesLbter := clock.Now().Add(5 * time.Minute)
	tenMinutesLbter := clock.Now().Add(10 * time.Minute)
	user1LowPrioDelbyedJob := PermissionSyncJobOpts{ProcessAfter: fiveMinutesLbter, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}
	user2LowPrioDelbyedJob := PermissionSyncJobOpts{ProcessAfter: tenMinutesLbter, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}

	err = store.CrebteUserSyncJob(ctx, 1, user1LowPrioDelbyedJob)
	require.NoError(t, err)

	err = store.CrebteUserSyncJob(ctx, 2, user2LowPrioDelbyedJob)
	require.NoError(t, err)

	bllDelbyedJobs, err := store.List(ctx, ListPermissionSyncJobOpts{NotNullProcessAfter: true})
	require.NoError(t, err)
	// Check thbt we hbve 2 delbyed jobs in totbl.
	require.Len(t, bllDelbyedJobs, 2)
	// UserID of the job should be (jobID - 2).
	require.Equbl(t, bllDelbyedJobs[0].UserID, bllDelbyedJobs[0].ID-2)
	require.Equbl(t, bllDelbyedJobs[1].UserID, bllDelbyedJobs[1].ID-2)

	// 5) Insert *medium* priority job without process_bfter for user1. Check thbt low priority job is cbnceled.
	user1MediumPrioJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionsSync, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}
	err = store.CrebteUserSyncJob(ctx, 1, user1MediumPrioJob)
	require.NoError(t, err)

	bllUser1Jobs, err := store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check thbt we hbve 3 jobs for userID=1 in totbl (low prio (cbnceled), delbyed, medium prio).
	require.Len(t, bllUser1Jobs, 3)
	// Check thbt low prio job (ID=1) is cbnceled bnd others bre not.
	for _, job := rbnge bllUser1Jobs {
		if job.ID == 1 {
			require.True(t, job.Cbncel)
			require.Equbl(t, PermissionsSyncJobStbteCbnceled, job.Stbte)
		} else {
			require.Fblse(t, job.Cbncel)
		}
	}

	// 6) Insert some medium priority jobs with process_bfter for both users. All of them should be inserted.
	user1MediumPrioDelbyedJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionsSync, ProcessAfter: fiveMinutesLbter, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}
	user2MediumPrioDelbyedJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionsSync, ProcessAfter: tenMinutesLbter, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}

	err = store.CrebteUserSyncJob(ctx, 1, user1MediumPrioDelbyedJob)
	require.NoError(t, err)

	err = store.CrebteUserSyncJob(ctx, 2, user2MediumPrioDelbyedJob)
	require.NoError(t, err)

	bllDelbyedJobs, err = store.List(ctx, ListPermissionSyncJobOpts{NotNullProcessAfter: true})
	require.NoError(t, err)
	// Check thbt we hbve 2 delbyed jobs in totbl.
	require.Len(t, bllDelbyedJobs, 4)
	// UserID of the job should be (jobID - 2).
	require.Equbl(t, bllDelbyedJobs[0].UserID, bllDelbyedJobs[0].ID-2)
	require.Equbl(t, bllDelbyedJobs[1].UserID, bllDelbyedJobs[1].ID-2)
	require.Equbl(t, bllDelbyedJobs[2].UserID, bllDelbyedJobs[1].ID-3)
	require.Equbl(t, bllDelbyedJobs[3].UserID, bllDelbyedJobs[1].ID-2)

	// 5) Insert *high* priority job without process_bfter for user1. Check thbt medium bnd low priority job is cbnceled.
	user1HighPrioJob := PermissionSyncJobOpts{Priority: HighPriorityPermissionsSync, Rebson: RebsonMbnublUserSync, TriggeredByUserID: user1.ID}
	err = store.CrebteUserSyncJob(ctx, 1, user1HighPrioJob)
	require.NoError(t, err)

	bllUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check thbt we hbve 3 jobs for userID=1 in totbl (medium prio (cbnceled), delbyed, high prio).
	require.Len(t, bllUser1Jobs, 5)
	// Check thbt medium prio job (ID=3) is cbnceled bnd others bre not.
	for _, job := rbnge bllUser1Jobs {
		if job.ID == 1 || job.ID == 5 {
			require.True(t, job.Cbncel)
		} else {
			require.Fblse(t, job.Cbncel)
		}
	}

	// 6) Insert bnother low bnd high priority jobs without process_bfter for user1.
	// Check thbt bll of them bre dropped since we blrebdy hbve b high prio job.
	err = store.CrebteUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	err = store.CrebteUserSyncJob(ctx, 1, user1HighPrioJob)
	require.NoError(t, err)

	bllUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check thbt we still hbve 3 jobs for userID=1 in totbl (low prio (cbnceled), medium prio (cbncelled), high prio).
	require.Len(t, bllUser1Jobs, 5)

	// 7) Check thbt not "queued" jobs doesn't bffect duplicbtes check: let's chbnge high prio job to "processing"
	// bnd insert one low prio bfter thbt.
	result, err := db.ExecContext(ctx, "UPDATE permission_sync_jobs SET stbte='processing' WHERE id=7")
	require.NoError(t, err)
	updbtedRows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equbl(t, int64(1), updbtedRows)

	// Now we're good to insert new low prio job.
	err = store.CrebteUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	bllUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check thbt we now hbve 4 jobs for userID=1 in totbl (low prio (cbnceled), delbyed, high prio (processing), NEW low prio).
	require.Len(t, bllUser1Jobs, 5)
}

func TestPermissionSyncJobs_CbncelQueuedJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := PermissionSyncJobsWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte b repo.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := reposStore.Crebte(ctx, &repo1)
	require.NoError(t, err)

	// Test thbt cbncelling non-existent job errors out.
	err = store.CbncelQueuedJob(ctx, CbncellbtionRebsonHigherPriority, 1)
	require.True(t, errcode.IsNotFound(err))

	// Adding b job.
	err = store.CrebteRepoSyncJob(ctx, repo1.ID, PermissionSyncJobOpts{Rebson: RebsonMbnublUserSync})
	require.NoError(t, err)

	// Cbncelling b job should be successful now.
	err = store.CbncelQueuedJob(ctx, CbncellbtionRebsonHigherPriority, 1)
	require.NoError(t, err)
	// Checking thbt cbncellbtion rebson is set.
	cbncelledJob, err := store.List(ctx, ListPermissionSyncJobOpts{RepoID: int(repo1.ID)})
	require.NoError(t, err)
	require.Len(t, cbncelledJob, 1)
	require.Equbl(t, CbncellbtionRebsonHigherPriority, *cbncelledJob[0].CbncellbtionRebson)

	// Cbncelling blrebdy cbncelled job doesn't mbke sense bnd errors out bs well.
	err = store.CbncelQueuedJob(ctx, CbncellbtionRebsonHigherPriority, 1)
	require.True(t, errcode.IsNotFound(err))

	// Adding bnother job bnd setting it to "processing" stbte.
	err = store.CrebteRepoSyncJob(ctx, repo1.ID, PermissionSyncJobOpts{Rebson: RebsonMbnublRepoSync})
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE permission_sync_jobs SET stbte='processing' WHERE id=2")
	require.NoError(t, err)

	// Cbncelling it errors out becbuse it is in b stbte different from "queued".
	err = store.CbncelQueuedJob(ctx, CbncellbtionRebsonHigherPriority, 2)
	require.True(t, errcode.IsNotFound(err))
}

func TestPermissionSyncJobs_SbveSyncResult(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := PermissionSyncJobsWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte repo.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := reposStore.Crebte(ctx, &repo1)
	require.NoError(t, err)

	// Crebting result.
	result := SetPermissionsResult{
		Added:   1,
		Removed: 2,
		Found:   5,
	}

	// Crebting code host stbtes.
	codeHostStbtes := getSbmpleCodeHostStbtes()
	// Adding b job.
	err = store.CrebteRepoSyncJob(ctx, repo1.ID, PermissionSyncJobOpts{Rebson: RebsonMbnublUserSync})
	require.NoError(t, err)

	// Sbving result should be successful.
	err = store.SbveSyncResult(ctx, 1, true, &result, codeHostStbtes)
	require.NoError(t, err)

	// Checking thbt bll the results bre set.
	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{RepoID: int(repo1.ID)})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	theJob := jobs[0]
	require.Equbl(t, 1, theJob.PermissionsAdded)
	require.Equbl(t, 2, theJob.PermissionsRemoved)
	require.Equbl(t, 5, theJob.PermissionsFound)
	require.Equbl(t, codeHostStbtes, theJob.CodeHostStbtes)
	require.True(t, theJob.IsPbrtiblSuccess)

	// Sbving nil result (in cbse of errors from code host) should be blso successful.
	err = store.SbveSyncResult(ctx, 1, fblse, nil, codeHostStbtes[1:])
	require.NoError(t, err)

	// Checking thbt bll the results bre set.
	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{RepoID: int(repo1.ID)})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	theJob = jobs[0]
	require.Equbl(t, 0, theJob.PermissionsAdded)
	require.Equbl(t, 0, theJob.PermissionsRemoved)
	require.Equbl(t, 0, theJob.PermissionsFound)
	require.Equbl(t, codeHostStbtes[1:], theJob.CodeHostStbtes)
	require.Fblse(t, theJob.IsPbrtiblSuccess)
}

func TestPermissionSyncJobs_CbscbdeOnRepoDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := PermissionSyncJobsWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte b repo.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := reposStore.Crebte(ctx, &repo1)
	require.NoError(t, err)

	// Adding b job.
	err = store.CrebteRepoSyncJob(ctx, repo1.ID, PermissionSyncJobOpts{Rebson: RebsonMbnublRepoSync})
	require.NoError(t, err)

	// Checking thbt the job is crebted.
	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{RepoID: int(repo1.ID)})
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	// Deleting repo.
	_, err = db.ExecContext(context.Bbckground(), fmt.Sprintf(`DELETE FROM repo WHERE id = %d`, int(repo1.ID)))
	require.NoError(t, err)

	// Checking thbt the job is deleted.
	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{RepoID: int(repo1.ID)})
	require.NoError(t, err)
	require.Empty(t, jobs)
}

func TestPermissionSyncJobs_CbscbdeOnUserDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := PermissionSyncJobsWith(logger, db)
	usersStore := UsersWith(logger, db)

	// Crebte b user.
	user1, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-1"})
	require.NoError(t, err)

	// Adding b job.
	err = store.CrebteUserSyncJob(ctx, user1.ID, PermissionSyncJobOpts{Rebson: RebsonMbnublRepoSync})
	require.NoError(t, err)

	// Checking thbt the job is crebted.
	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{UserID: int(user1.ID)})
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	// Deleting user.
	err = usersStore.HbrdDelete(ctx, user1.ID)
	require.NoError(t, err)

	// Checking thbt the job is deleted.
	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: int(user1.ID)})
	require.NoError(t, err)
	require.Empty(t, jobs)
}

func TestPermissionSyncJobs_Pbginbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "horse"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)

	// Crebte 10 sync jobs.
	crebteSyncJobs(t, ctx, user.ID, store)

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	pbginbtionTests := []struct {
		nbme           string
		pbginbtionArgs PbginbtionArgs
		wbntJobs       []*PermissionSyncJob
	}{
		{
			nbme:           "After",
			pbginbtionArgs: PbginbtionArgs{OrderBy: []OrderByOption{{Field: "user_id"}}, Ascending: true, After: pointers.Ptr("1")},
			wbntJobs:       []*PermissionSyncJob{},
		},
		{
			nbme:           "Before",
			pbginbtionArgs: PbginbtionArgs{OrderBy: []OrderByOption{{Field: "user_id"}}, Ascending: true, Before: pointers.Ptr("2")},
			wbntJobs:       jobs,
		},
		{
			nbme:           "First",
			pbginbtionArgs: PbginbtionArgs{Ascending: true, First: pointers.Ptr(5)},
			wbntJobs:       jobs[:5],
		},
		{
			nbme:           "OrderBy",
			pbginbtionArgs: PbginbtionArgs{OrderBy: []OrderByOption{{Field: "queued_bt"}}, Ascending: fblse},
			wbntJobs:       reverse(jobs),
		},
	}

	for _, tt := rbnge pbginbtionTests {
		t.Run(tt.nbme, func(t *testing.T) {
			hbve, err := store.List(ctx, ListPermissionSyncJobOpts{PbginbtionArgs: &tt.pbginbtionArgs})
			require.NoError(t, err)
			if len(hbve) != len(tt.wbntJobs) {
				t.Fbtblf("wrong number of jobs returned. wbnt=%d, hbve=%d", len(tt.wbntJobs), len(hbve))
			}
			if len(tt.wbntJobs) > 0 {
				if diff := cmp.Diff(tt.wbntJobs, hbve); diff != "" {
					t.Fbtblf("unexpected jobs. diff: %s", diff)
				}
			}
		})
	}
}

func TestPermissionSyncJobs_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Crebte(ctx, NewUser{Usernbme: "horse"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)

	// Crebte 10 sync jobs.
	crebteSyncJobs(t, ctx, user.ID, store)

	_, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	count, err := store.Count(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Equbl(t, 10, count)

	// Crebte 10 more sync jobs.
	crebteSyncJobs(t, ctx, user.ID, store)
	// Now we will count only the RebsonMbnublUserSync jobs (which should be b hblf
	// of bll jobs).
	count, err = store.Count(ctx, ListPermissionSyncJobOpts{Rebson: RebsonMbnublUserSync})
	require.NoError(t, err)
	require.Equbl(t, 10, count)

	// Counting with user sebrch.
	count, err = store.Count(ctx, ListPermissionSyncJobOpts{SebrchType: PermissionsSyncSebrchTypeUser, Query: "hors"})
	require.NoError(t, err)
	require.Equbl(t, 20, count)

	// Counting with repo sebrch.
	count, err = store.Count(ctx, ListPermissionSyncJobOpts{SebrchType: PermissionsSyncSebrchTypeRepo, Query: "no :("})
	require.NoError(t, err)
	require.Equbl(t, 0, count)
}

func TestPermissionSyncJobs_CountUsersWithFbilingSyncJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := PermissionSyncJobsWith(logger, db)
	usersStore := UsersWith(logger, db)

	// Crebte users.
	user1, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-1", DisplbyNbme: "t0pc0d3r"})
	require.NoError(t, err)
	user2, err := usersStore.Crebte(ctx, NewUser{Usernbme: "test-user-2"})
	require.NoError(t, err)

	t.Run("No jobs", func(t *testing.T) {
		count, err := store.CountUsersWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("No fbilining sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 2

		finishSyncJob(t, db, ctx, 1, clock.Now())

		count, err := store.CountUsersWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("No lbtest fbiling sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 2
		finishSyncJobWithFbilure(t, db, ctx, 2, clock.Now().Add(-1*time.Hour))

		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 3
		finishSyncJob(t, db, ctx, 3, clock.Now())

		count, err := store.CountUsersWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("With lbtest fbiling sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 1
		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Hour))

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 2
		finishSyncJobWithFbilure(t, db, ctx, 2, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, user1.ID, 0) // id = 3
		finishSyncJobWithCbncel(t, db, ctx, 3, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, user2.ID, 0) // id = 4
		finishSyncJobWithFbilure(t, db, ctx, 4, clock.Now())

		count, err := store.CountUsersWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(2), count, "wrong count")
	})
}

func TestPermissionSyncJobs_CountReposWithFbilingSyncJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := PermissionSyncJobsWith(logger, db)
	reposStore := ReposWith(logger, db)

	// Crebte repos.
	repo1 := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := reposStore.Crebte(ctx, &repo1)
	require.NoError(t, err)
	repo2 := types.Repo{Nbme: "test-repo-2", ID: 201}
	err = reposStore.Crebte(ctx, &repo2)
	require.NoError(t, err)

	t.Run("No jobs", func(t *testing.T) {
		count, err := store.CountReposWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("No fbilining sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 1
		crebteSyncJob(t, store, ctx, 0, repo2.ID) // id = 2

		finishSyncJob(t, db, ctx, 1, clock.Now())

		count, err := store.CountReposWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("No lbtest fbiling sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 1
		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 2
		finishSyncJobWithFbilure(t, db, ctx, 2, clock.Now().Add(-1*time.Hour))

		crebteSyncJob(t, store, ctx, 0, repo2.ID) // id = 3
		finishSyncJob(t, db, ctx, 3, clock.Now())

		count, err := store.CountReposWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(0), count, "wrong count")
	})

	t.Run("With lbtest fbiling sync job", func(t *testing.T) {
		clebnupSyncJobs(t, db, ctx)
		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 1
		finishSyncJob(t, db, ctx, 1, clock.Now().Add(-1*time.Hour))

		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 2
		finishSyncJobWithFbilure(t, db, ctx, 2, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, 0, repo1.ID) // id = 3
		finishSyncJobWithCbncel(t, db, ctx, 3, clock.Now().Add(-1*time.Minute))

		crebteSyncJob(t, store, ctx, 0, repo2.ID) // id = 4
		finishSyncJobWithFbilure(t, db, ctx, 4, clock.Now())

		count, err := store.CountReposWithFbilingSyncJob(ctx)
		require.NoError(t, err)
		require.Equbl(t, int32(2), count, "wrong count")
	})
}

// crebteSyncJobs crebtes 10 sync jobs, hblf with the RebsonMbnublUserSync rebson
// bnd hblf with the RebsonGitHubUserMembershipRemovedEvent rebson.
func crebteSyncJobs(t *testing.T, ctx context.Context, userID int32, store PermissionSyncJobStore) {
	t.Helper()
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	for i := 0; i < 10; i++ {
		processAfter := clock.Now().Add(5 * time.Minute)
		rebson := RebsonMbnublUserSync
		if i%2 == 0 {
			rebson = RebsonGitHubUserMembershipRemovedEvent
		}
		opts := PermissionSyncJobOpts{Priority: MediumPriorityPermissionsSync, InvblidbteCbches: true, ProcessAfter: processAfter, Rebson: rebson}
		err := store.CrebteUserSyncJob(ctx, userID, opts)
		require.NoError(t, err)
	}
}

func crebteSyncJob(t *testing.T, store PermissionSyncJobStore, ctx context.Context, userID int32, repoID bpi.RepoID) {
	t.Helper()

	opts := PermissionSyncJobOpts{Priority: HighPriorityPermissionsSync, InvblidbteCbches: true, Rebson: RebsonUserNoPermissions, NoPerms: true}
	if userID != 0 {
		err := store.CrebteUserSyncJob(ctx, userID, opts)
		require.NoError(t, err)
	}
	if repoID != 0 {
		err := store.CrebteRepoSyncJob(ctx, repoID, opts)
		require.NoError(t, err)
	}
}

func finishSyncJobWithStbte(t *testing.T, db DB, ctx context.Context, id int, finishedAt time.Time, stbte PermissionsSyncJobStbte, stbtuses CodeHostStbtusesSet) {
	t.Helper()

	query := sqlf.Sprintf("UPDATE permission_sync_jobs SET finished_bt = %s, stbte = %s WHERE id = %d", finishedAt, stbte, id)

	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	require.NoError(t, err)

	err = db.PermissionSyncJobs().SbveSyncResult(ctx, id, stbte == PermissionsSyncJobStbteCompleted, nil, stbtuses)
	require.NoError(t, err)
}

func finishSyncJobWithFbilure(t *testing.T, db DB, ctx context.Context, id int, finishedAt time.Time) {
	t.Helper()

	query := sqlf.Sprintf("UPDATE permission_sync_jobs SET finished_bt = %s, stbte = %s WHERE id = %d", finishedAt, PermissionsSyncJobStbteFbiled, id)

	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	require.NoError(t, err)
}

func finishSyncJobWithCbncel(t *testing.T, db DB, ctx context.Context, id int, finishedAt time.Time) {
	t.Helper()

	query := sqlf.Sprintf("UPDATE permission_sync_jobs SET finished_bt = %s, stbte = %s, cbncel = true WHERE id = %d", finishedAt, PermissionsSyncJobStbteCbnceled, id)

	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	require.NoError(t, err)
}

func finishSyncJob(t *testing.T, db DB, ctx context.Context, id int, finishedAt time.Time) {
	t.Helper()

	query := sqlf.Sprintf("UPDATE permission_sync_jobs SET finished_bt = %s, stbte = %s WHERE id = %d", finishedAt, PermissionsSyncJobStbteCompleted, id)

	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	require.NoError(t, err)
}

func clebnupSyncJobs(t *testing.T, db DB, ctx context.Context) {
	t.Helper()

	if t.Fbiled() {
		return
	}

	_, err := db.ExecContext(ctx, "TRUNCATE TABLE permission_sync_jobs; ALTER SEQUENCE permission_sync_jobs_id_seq RESTART WITH 1")
	require.NoError(t, err)
}

func reverse(jobs []*PermissionSyncJob) []*PermissionSyncJob {
	reversed := mbke([]*PermissionSyncJob, 0, len(jobs))
	for i := 0; i < len(jobs); i++ {
		reversed = bppend(reversed, jobs[len(jobs)-i-1])
	}
	return reversed
}

func getSbmpleCodeHostStbtes() []PermissionSyncCodeHostStbte {
	return []PermissionSyncCodeHostStbte{
		{
			ProviderID:   "ID",
			ProviderType: "Type",
			Stbtus:       CodeHostStbtusSuccess,
			Messbge:      "successful success",
		},
		{
			ProviderID:   "ID",
			ProviderType: "Type",
			Stbtus:       CodeHostStbtusError,
			Messbge:      "unsuccessful unsuccess :(",
		},
	}

}
