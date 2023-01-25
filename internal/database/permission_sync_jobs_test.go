package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/stretchr/testify/require"
)

const (
	// ReasonManualRepoSync and ReasonManualUserSync are copied from permssync
	// package to avoid import cycles.
	ReasonManualRepoSync = "REASON_MANUAL_REPO_SYNC"
	ReasonManualUserSync = "REASON_MANUAL_USER_SYNC"
)

func TestPermissionSyncJobs_CreateAndList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFakeClock(time.Now(), 0)

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user, err := db.Users().Create(ctx, NewUser{Username: "horse"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)

	jobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 0, "jobs returned even though database is empty")

	opts := PermissionSyncJobOpts{Priority: HighPriorityPermissionSync, InvalidateCaches: true, Reason: ReasonManualRepoSync, TriggeredByUserID: user.ID}
	err = store.CreateRepoSyncJob(ctx, 99, opts)
	require.NoError(t, err)

	processAfter := clock.Now().Add(5 * time.Minute)
	opts = PermissionSyncJobOpts{Priority: MediumPriorityPermissionSync, InvalidateCaches: true, ProcessAfter: processAfter, Reason: ReasonManualUserSync}
	err = store.CreateUserSyncJob(ctx, 77, opts)
	require.NoError(t, err)

	processAfter = clock.Now().Add(5 * time.Minute)
	opts = PermissionSyncJobOpts{Priority: LowPriorityPermissionSync, InvalidateCaches: true, ProcessAfter: processAfter, Reason: ReasonManualUserSync}
	err = store.CreateUserSyncJob(ctx, 78, opts)
	require.NoError(t, err)

	jobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)

	require.Len(t, jobs, 3, "wrong number of jobs returned")

	wantJobs := []*PermissionSyncJob{
		{
			ID:                jobs[0].ID,
			State:             "queued",
			RepositoryID:      99,
			Priority:          HighPriorityPermissionSync,
			InvalidateCaches:  true,
			Reason:            ReasonManualRepoSync,
			TriggeredByUserID: user.ID,
		},
		{
			ID:               jobs[1].ID,
			State:            "queued",
			UserID:           77,
			Priority:         MediumPriorityPermissionSync,
			InvalidateCaches: true,
			ProcessAfter:     processAfter,
			Reason:           ReasonManualUserSync,
		},
		{
			ID:               jobs[2].ID,
			State:            "queued",
			UserID:           78,
			Priority:         LowPriorityPermissionSync,
			InvalidateCaches: true,
			ProcessAfter:     processAfter,
			Reason:           ReasonManualUserSync,
		},
	}
	if diff := cmp.Diff(jobs, wantJobs, cmpopts.IgnoreFields(PermissionSyncJob{}, "QueuedAt")); diff != "" {
		t.Fatalf("jobs[0] has wrong attributes: %s", diff)
	}
	for i, j := range jobs {
		require.NotZerof(t, j.QueuedAt, "job %d has no QueuedAt set", i)
	}

	listTests := []struct {
		name     string
		opts     ListPermissionSyncJobOpts
		wantJobs []*PermissionSyncJob
	}{
		{
			name:     "ID",
			opts:     ListPermissionSyncJobOpts{ID: jobs[0].ID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "RepoID",
			opts:     ListPermissionSyncJobOpts{RepoID: jobs[0].RepositoryID},
			wantJobs: jobs[0:1],
		},
		{
			name:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[1].UserID},
			wantJobs: jobs[1:2],
		},
		{
			name:     "UserID",
			opts:     ListPermissionSyncJobOpts{UserID: jobs[2].UserID},
			wantJobs: jobs[2:],
		},
	}

	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := store.List(ctx, tt.opts)
			require.NoError(t, err)
			if len(have) != len(tt.wantJobs) {
				t.Fatalf("wrong number of jobs returned. want=%d, have=%d", len(tt.wantJobs), len(have))
			}
			if diff := cmp.Diff(have, tt.wantJobs); diff != "" {
				t.Fatalf("unexpected jobs. diff: %s", diff)
			}
		})
	}
}

func TestPermissionSyncJobs_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := timeutil.NewFakeClock(time.Now(), 0)

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	user1, err := db.Users().Create(ctx, NewUser{Username: "horse"})
	require.NoError(t, err)

	user2, err := db.Users().Create(ctx, NewUser{Username: "graph"})
	require.NoError(t, err)

	store := PermissionSyncJobsWith(logger, db)

	// 1) Insert low priority job without process_after for user1.
	user1LowPrioJob := PermissionSyncJobOpts{Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	allJobs, err := store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check that we have 1 job with userID=1.
	require.Len(t, allJobs, 1)
	require.Equal(t, 1, allJobs[0].UserID)

	// 2) Insert low priority job without process_after for user2.
	user2LowPrioJob := PermissionSyncJobOpts{Reason: ReasonManualUserSync, TriggeredByUserID: user2.ID}
	err = store.CreateUserSyncJob(ctx, 2, user2LowPrioJob)
	require.NoError(t, err)

	allJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check that we have 2 jobs including job for userID=2. Job ID should match user ID.
	require.Len(t, allJobs, 2)
	require.Equal(t, allJobs[0].ID, allJobs[0].UserID)
	require.Equal(t, allJobs[1].ID, allJobs[1].UserID)

	// 3) Another low priority job without process_after for user1 is dropped.
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	allJobs, err = store.List(ctx, ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	// Check that we still have 2 jobs. Job ID should match user ID.
	require.Len(t, allJobs, 2)
	require.Equal(t, allJobs[0].ID, allJobs[0].UserID)
	require.Equal(t, allJobs[1].ID, allJobs[1].UserID)

	// 4) Insert some low priority jobs with process_after for both users. All of them should be inserted.
	fiveMinutesLater := clock.Now().Add(5 * time.Minute)
	tenMinutesLater := clock.Now().Add(10 * time.Minute)
	user1LowPrioDelayedJob := PermissionSyncJobOpts{ProcessAfter: fiveMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	user2LowPrioDelayedJob := PermissionSyncJobOpts{ProcessAfter: tenMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}

	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioDelayedJob)
	require.NoError(t, err)

	err = store.CreateUserSyncJob(ctx, 2, user2LowPrioDelayedJob)
	require.NoError(t, err)

	allDelayedJobs, err := store.List(ctx, ListPermissionSyncJobOpts{NotNullProcessAfter: true})
	require.NoError(t, err)
	// Check that we have 2 delayed jobs in total.
	require.Len(t, allDelayedJobs, 2)
	// UserID of the job should be (jobID - 2).
	require.Equal(t, allDelayedJobs[0].UserID, allDelayedJobs[0].ID-2)
	require.Equal(t, allDelayedJobs[1].UserID, allDelayedJobs[1].ID-2)

	// 5) Insert *medium* priority job without process_after for user1. Check that low priority job is canceled.
	user1MediumPrioJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionSync, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1MediumPrioJob)
	require.NoError(t, err)

	allUser1Jobs, err := store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check that we have 3 jobs for userID=1 in total (low prio (canceled), delayed, medium prio).
	require.Len(t, allUser1Jobs, 3)
	// Check that low prio job (ID=1) is canceled and others are not.
	for _, job := range allUser1Jobs {
		if job.ID == 1 {
			require.True(t, job.Cancel)
		} else {
			require.False(t, job.Cancel)
		}
	}

	// 6) Insert some medium priority jobs with process_after for both users. All of them should be inserted.
	user1MediumPrioDelayedJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionSync, ProcessAfter: fiveMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	user2MediumPrioDelayedJob := PermissionSyncJobOpts{Priority: MediumPriorityPermissionSync, ProcessAfter: tenMinutesLater, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}

	err = store.CreateUserSyncJob(ctx, 1, user1MediumPrioDelayedJob)
	require.NoError(t, err)

	err = store.CreateUserSyncJob(ctx, 2, user2MediumPrioDelayedJob)
	require.NoError(t, err)

	allDelayedJobs, err = store.List(ctx, ListPermissionSyncJobOpts{NotNullProcessAfter: true})
	require.NoError(t, err)
	// Check that we have 2 delayed jobs in total.
	require.Len(t, allDelayedJobs, 4)
	// UserID of the job should be (jobID - 2).
	require.Equal(t, allDelayedJobs[0].UserID, allDelayedJobs[0].ID-2)
	require.Equal(t, allDelayedJobs[1].UserID, allDelayedJobs[1].ID-2)
	require.Equal(t, allDelayedJobs[2].UserID, allDelayedJobs[1].ID-3)
	require.Equal(t, allDelayedJobs[3].UserID, allDelayedJobs[1].ID-2)

	// 5) Insert *high* priority job without process_after for user1. Check that medium and low priority job is canceled.
	user1HighPrioJob := PermissionSyncJobOpts{Priority: HighPriorityPermissionSync, Reason: ReasonManualUserSync, TriggeredByUserID: user1.ID}
	err = store.CreateUserSyncJob(ctx, 1, user1HighPrioJob)
	require.NoError(t, err)

	allUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check that we have 3 jobs for userID=1 in total (medium prio (canceled), delayed, high prio).
	require.Len(t, allUser1Jobs, 5)
	// Check that medium prio job (ID=3) is canceled and others are not.
	for _, job := range allUser1Jobs {
		if job.ID == 1 || job.ID == 5 {
			require.True(t, job.Cancel)
		} else {
			require.False(t, job.Cancel)
		}
	}

	// 6) Insert another low and high priority jobs without process_after for user1.
	// Check that all of them are dropped since we already have a high prio job.
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	err = store.CreateUserSyncJob(ctx, 1, user1HighPrioJob)
	require.NoError(t, err)

	allUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check that we still have 3 jobs for userID=1 in total (low prio (canceled), medium prio (cancelled), high prio).
	require.Len(t, allUser1Jobs, 5)

	// 7) Check that not "queued" jobs doesn't affect duplicates check: let's change high prio job to "processing"
	// and insert one low prio after that.
	result, err := db.ExecContext(ctx, "UPDATE permission_sync_jobs SET state='processing' WHERE id=7")
	require.NoError(t, err)
	updatedRows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), updatedRows)

	// Now we're good to insert new low prio job.
	err = store.CreateUserSyncJob(ctx, 1, user1LowPrioJob)
	require.NoError(t, err)

	allUser1Jobs, err = store.List(ctx, ListPermissionSyncJobOpts{UserID: 1})
	require.NoError(t, err)
	// Check that we now have 4 jobs for userID=1 in total (low prio (canceled), delayed, high prio (processing), NEW low prio).
	require.Len(t, allUser1Jobs, 5)
}

func TestPermissionSyncJobs_CancelQueuedJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	// Test that cancelling non-existent job errors out.
	err := store.CancelQueuedJob(ctx, CancellationReasonHigherPriority, 1)
	require.True(t, errcode.IsNotFound(err))

	// Adding a job.
	err = store.CreateRepoSyncJob(ctx, 1, PermissionSyncJobOpts{Reason: ReasonManualUserSync})
	require.NoError(t, err)

	// Cancelling a job should be successful now.
	err = store.CancelQueuedJob(ctx, CancellationReasonHigherPriority, 1)
	require.NoError(t, err)
	// Checking that cancellation reason is set.
	cancelledJob, err := store.List(ctx, ListPermissionSyncJobOpts{RepoID: 1})
	require.NoError(t, err)
	require.Len(t, cancelledJob, 1)
	require.Equal(t, CancellationReasonHigherPriority, cancelledJob[0].CancellationReason)

	// Cancelling already cancelled job doesn't make sense and errors out as well.
	err = store.CancelQueuedJob(ctx, CancellationReasonHigherPriority, 1)
	require.True(t, errcode.IsNotFound(err))

	// Adding another job and setting it to "processing" state.
	err = store.CreateRepoSyncJob(ctx, 1, PermissionSyncJobOpts{Reason: ReasonManualUserSync})
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE permission_sync_jobs SET state='processing' WHERE id=2")
	require.NoError(t, err)

	// Cancelling it errors out because it is in a state different from "queued".
	err = store.CancelQueuedJob(ctx, CancellationReasonHigherPriority, 2)
	require.True(t, errcode.IsNotFound(err))
}
