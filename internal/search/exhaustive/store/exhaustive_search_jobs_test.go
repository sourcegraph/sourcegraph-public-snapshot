package store_test

import (
	"context"
	"testing"

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

	t.Cleanup(func() {
		cleanupUsers(bs)
	})

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
			expectedErr: errors.New("ERROR: duplicate key value violates unique constraint \"exhaustive_search_jobs_query_initiator_id_key\" (SQLSTATE 23505)"),
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
			t.Cleanup(func() {
				cleanupSearchJobs(bs)
			})

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
