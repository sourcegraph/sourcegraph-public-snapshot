package store_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

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

		// TODO(keegancsmith) for some reason we don't let users recreate searches.
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
			expectedErr: nil,
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	bs := basestore.NewWithHandle(db.Handle())

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)

	ctx := actor.WithActor(context.Background(), actor.FromUser(userID))

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
	haveJobs, err := s.ListExhaustiveSearchJobs(ctx)
	require.NoError(t, err)
	require.Equal(t, len(haveJobs), len(jobs))

	haveIDs := make([]int64, len(haveJobs))
	for i, job := range haveJobs {
		haveIDs[i] = job.ID
	}
	wantIDs := make([]int64, len(jobs))
	for i, job := range jobs {
		wantIDs[i] = job.ID
	}

	if diff := cmp.Diff(haveIDs, wantIDs); diff != "" {
		t.Fatalf("List returned wrong jobs: %s", diff)
	}
}
