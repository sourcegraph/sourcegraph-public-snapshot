package store_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestStore_CreateExhaustiveSearchJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	bs := basestore.NewWithHandle(db.Handle())

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)
	malloryID, err := createUser(bs, "mallory")
	require.NoError(t, err)
	adminID, err := createUser(bs, "admin")
	require.NoError(t, err)

	s := store.New(db, &observation.TestContext)

	tests := []struct {
		name        string
		setup       func(context.Context, *store.Store) error
		job         types.ExhaustiveSearchJob
		actor       *actor.Actor // defaults to "alice"
		expectedErr error
	}{
		{
			name: "New job",
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchJob",
			},
			expectedErr: nil,
		},
		{
			name: "Missing user ID",
			job: types.ExhaustiveSearchJob{
				Query: "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchJob",
			},
			expectedErr: errors.New("missing initiator ID"),
		},
		{
			name: "Missing query",
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
			},
			expectedErr: errors.New("missing query"),
		},

		{
			name: "Search already exists",
			setup: func(ctx context.Context, s *store.Store) error {
				_, err := s.CreateExhaustiveSearchJob(ctx, types.ExhaustiveSearchJob{
					InitiatorID: userID,
					Query:       "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchJob_exists",
				})
				return err
			},
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchJob_exists",
			},
		},

		// Security tests
		{
			name: "admin can spoof",
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "fear me",
			},
			actor:       &actor.Actor{UID: adminID},
			expectedErr: nil,
		},
		{
			name: "malicious user cant spoof",
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "the cake is a lie",
			},
			actor:       &actor.Actor{UID: malloryID},
			expectedErr: auth.ErrMustBeSiteAdminOrSameUser,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act := test.actor
			if act == nil {
				act = &actor.Actor{UID: userID}
			}
			ctx := actor.WithActor(context.Background(), act)

			if test.setup != nil {
				require.NoError(t, test.setup(ctx, s))
			}

			jobID, err := s.CreateExhaustiveSearchJob(ctx, test.job)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.NotZero(t, jobID)
			}
		})
	}
}

func TestStore_GetAndListSearchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	bs := basestore.NewWithHandle(db.Handle())

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)

	adminID, err := createUser(bs, "admin")
	require.NoError(t, err)

	ctx := actor.WithActor(context.Background(), actor.FromUser(userID))
	adminCtx := actor.WithActor(context.Background(), actor.FromUser(adminID))

	s := store.New(db, &observation.TestContext)

	jobs := []types.ExhaustiveSearchJob{
		{InitiatorID: userID, Query: "repo:job1"},
		{InitiatorID: userID, Query: "repo:job2"},
		{InitiatorID: userID, Query: "repo:job3"},
	}

	// Create jobs
	for i, job := range jobs {
		jobID, err := s.CreateExhaustiveSearchJob(ctx, job)
		require.NoError(t, err)
		assert.NotZero(t, jobID)

		jobs[i].ID = jobID
	}

	// Now get them one-by-one
	for _, job := range jobs {
		haveJob, err := s.GetExhaustiveSearchJob(ctx, job.ID)
		require.NoError(t, err)

		// Ensure we got the right job and that the fields are scanned correctly
		assert.Equal(t, haveJob.ID, job.ID)
		assert.Equal(t, haveJob.Query, job.Query)
		assert.Equal(t, haveJob.State, types.JobStateQueued)
		assert.NotZero(t, haveJob.CreatedAt)
		assert.NotZero(t, haveJob.UpdatedAt)
	}

	// Now list them all

	tc := []struct {
		name    string
		ctx     context.Context
		args    store.ListArgs
		wantIDs []int64
		wantErr bool
	}{
		{
			name: "query: 1 job",
			ctx:  ctx,
			args: store.ListArgs{
				Query: "job1",
			},
			wantIDs: []int64{jobs[0].ID},
		},
		{
			name: "query: all jobs",
			ctx:  ctx,
			args: store.ListArgs{
				Query: "repo",
			},
			wantIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			name: "states: queued jobs",
			ctx:  ctx,
			args: store.ListArgs{
				States: []string{string(types.JobStateQueued)},
			},
			wantIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			name: "query: all jobs but ask for 1 job only",
			ctx:  ctx,
			args: store.ListArgs{
				PaginationArgs: &database.PaginationArgs{First: intptr(1), Ascending: true},
				Query:          "repo",
			},
			wantIDs: []int64{jobs[0].ID},
		},
		// negative test
		{
			name: "query: no result",
			ctx:  ctx,
			args: store.ListArgs{
				Query: "foo",
			},
			wantIDs: []int64{},
		},
		{
			name: "states: no result",
			ctx:  ctx,
			args: store.ListArgs{
				States: []string{string(types.JobStateCompleted)},
			},
			wantIDs: []int64{},
		},
		// Security tests
		{
			name: "userIDs: Admins can ask for userIDs",
			ctx:  adminCtx,
			args: store.ListArgs{
				UserIDs: []int32{userID},
			},
			wantIDs: []int64{jobs[0].ID, jobs[1].ID, jobs[2].ID},
		},
		{
			name: "userIDs: Non-admins CANNOT ask for userIDs",
			ctx:  ctx,
			args: store.ListArgs{
				UserIDs: []int32{userID + 1},
			},
			wantIDs: []int64{},
			wantErr: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			haveJobs, err := s.ListExhaustiveSearchJobs(c.ctx, c.args)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, len(haveJobs), len(c.wantIDs))

			haveIDs := make([]int64, len(haveJobs))
			for i, job := range haveJobs {
				haveIDs[i] = job.ID
			}

			if diff := cmp.Diff(haveIDs, c.wantIDs); diff != "" {
				t.Fatalf("List returned wrong jobs: %s", diff)
			}
		})
	}
}

// TestStore_GetAggregateStatus tests that ListExhaustiveSearchJobs returns the
// proper aggregated state.
func TestStore_AggregateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	bs := basestore.NewWithHandle(db.Handle())

	_, err := createRepo(db, "repo1")
	require.NoError(t, err)

	s := store.New(db, &observation.TestContext)

	tc := []struct {
		name string
		c    stateCascade
		want types.JobState
	}{
		{
			name: "only repo rev jobs running",
			c: stateCascade{
				searchJob:   types.JobStateCompleted,
				repoJobs:    []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{types.JobStateProcessing},
			},
			want: types.JobStateProcessing,
		},
		{
			name: "processing, because at least 1 job is running",
			c: stateCascade{
				searchJob: types.JobStateProcessing,
				repoJobs:  []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{
					types.JobStateProcessing,
					types.JobStateQueued,
					types.JobStateCompleted,
				},
			},
			want: types.JobStateProcessing,
		},
		{
			name: "processing, although some jobs failed",
			c: stateCascade{
				searchJob: types.JobStateCompleted,
				repoJobs:  []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{
					types.JobStateProcessing,
					types.JobStateFailed,
				},
			},
			want: types.JobStateProcessing,
		},
		{
			name: "all jobs finished, at least 1 failed",
			c: stateCascade{
				searchJob:   types.JobStateCompleted,
				repoJobs:    []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{types.JobStateCompleted, types.JobStateFailed},
			},
			want: types.JobStateFailed,
		},
		{
			name: "all jobs finished successfully",
			c: stateCascade{
				searchJob:   types.JobStateCompleted,
				repoJobs:    []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{types.JobStateCompleted, types.JobStateCompleted},
			},
			want: types.JobStateCompleted,
		},
		{
			name: "search job was canceled, but some jobs haven't stopped yet",
			c: stateCascade{
				searchJob:   types.JobStateCanceled,
				repoJobs:    []types.JobState{types.JobStateCompleted},
				repoRevJobs: []types.JobState{types.JobStateProcessing, types.JobStateFailed},
			},
			want: types.JobStateCanceled,
		},
		{
			name: "top-level search job finished, but the other jobs haven't started yet",
			c: stateCascade{
				searchJob: types.JobStateCompleted,
				repoJobs:  []types.JobState{types.JobStateQueued},
			},
			want: types.JobStateQueued,
		},
		{
			name: "search job is queued, but no other job has been created yet",
			c: stateCascade{
				searchJob: types.JobStateQueued,
			},
			want: types.JobStateQueued,
		},
	}

	for i, tt := range tc {
		t.Run("", func(t *testing.T) {
			userID, err := createUser(bs, fmt.Sprintf("user_%d", i))
			require.NoError(t, err)

			ctx := actor.WithActor(context.Background(), actor.FromUser(userID))
			jobID := createJobCascade(t, ctx, s, tt.c)

			jobs, err := s.ListExhaustiveSearchJobs(ctx, store.ListArgs{})
			require.NoError(t, err)
			require.Equal(t, 1, len(jobs))
			require.Equal(t, jobID, jobs[0].ID)
			assert.Equal(t, tt.want, jobs[0].AggState)
		})
	}
}

// createJobCascade creates a cascade of jobs (1 search job -> n repo jobs -> m
// repo rev jobs) with states as defined in stateCascade.
//
// This is a fairly large test helper, because don't want to start the worker
// routines, but instead we want to create a snapshot of the state of the jobs
// at a given point in time.
func createJobCascade(
	t *testing.T,
	ctx context.Context,
	stor *store.Store,
	casc stateCascade,
) (searchJobID int64) {
	t.Helper()

	searchJob := types.ExhaustiveSearchJob{
		InitiatorID: actor.FromContext(ctx).UID,
		Query:       "repo:job1",
		WorkerJob:   types.WorkerJob{State: casc.searchJob},
	}

	repoJobs := make([]types.ExhaustiveSearchRepoJob, len(casc.repoJobs))
	for i, r := range casc.repoJobs {
		repoJobs[i] = types.ExhaustiveSearchRepoJob{
			WorkerJob: types.WorkerJob{State: r},
			RepoID:    1, // same repo for all tests
			RefSpec:   "HEAD",
		}
	}

	repoRevJobs := make([]types.ExhaustiveSearchRepoRevisionJob, len(casc.repoRevJobs))
	for i, rr := range casc.repoRevJobs {
		repoRevJobs[i] = types.ExhaustiveSearchRepoRevisionJob{
			WorkerJob: types.WorkerJob{State: rr},
			Revision:  "HEAD",
		}
	}

	jobID, err := stor.CreateExhaustiveSearchJob(ctx, searchJob)
	require.NoError(t, err)
	assert.NotZero(t, jobID)

	err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_jobs SET state = %s WHERE id = %s", casc.searchJob, jobID))
	require.NoError(t, err)

	for i, r := range repoJobs {
		r.SearchJobID = jobID
		repoJobID, err := stor.CreateExhaustiveSearchRepoJob(ctx, r)
		require.NoError(t, err)
		assert.NotZero(t, repoJobID)

		err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_repo_jobs SET state = %s WHERE id = %s", casc.repoJobs[i], repoJobID))
		require.NoError(t, err)

		for j, rr := range repoRevJobs {
			rr.SearchRepoJobID = repoJobID
			repoRevJobID, err := stor.CreateExhaustiveSearchRepoRevisionJob(ctx, rr)
			require.NoError(t, err)
			assert.NotZero(t, repoRevJobID)
			require.NoError(t, err)

			err = stor.Exec(ctx, sqlf.Sprintf("UPDATE exhaustive_search_repo_revision_jobs SET state = %s WHERE id = %s", casc.repoRevJobs[j], repoRevJobID))
			require.NoError(t, err)
		}
	}

	return jobID
}

type stateCascade struct {
	searchJob   types.JobState
	repoJobs    []types.JobState
	repoRevJobs []types.JobState
}

func intptr(s int) *int { return &s }
