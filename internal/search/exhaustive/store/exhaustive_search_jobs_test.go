package store_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestStore_CreateExhaustiveSearchJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bs := basestore.NewWithHandle(db.Handle())

	t.Cleanup(func() {
		cleanupUsers(bs)
	})

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)

	s := store.New(db, &observation.TestContext)

	tests := []struct {
		name        string
		setup       func(*testing.T, *store.Store)
		job         types.ExhaustiveSearchJob
		expectedErr error
	}{
		{
			name: "New job",
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "repo:^github\\.com/hashicorp/errwrap$ hello",
			},
			expectedErr: nil,
		},
		{
			name: "Missing user ID",
			job: types.ExhaustiveSearchJob{
				Query: "repo:^github\\.com/hashicorp/errwrap$ hello",
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
			setup: func(t *testing.T, s *store.Store) {
				_, err := s.CreateExhaustiveSearchJob(context.Background(), types.ExhaustiveSearchJob{
					InitiatorID: userID,
					Query:       "repo:^github\\.com/hashicorp/errwrap$ hello",
				})
				require.NoError(t, err)
			},
			job: types.ExhaustiveSearchJob{
				InitiatorID: userID,
				Query:       "repo:^github\\.com/hashicorp/errwrap$ hello",
			},
			expectedErr: errors.New("ERROR: duplicate key value violates unique constraint \"exhaustive_search_jobs_query_initiator_id_key\" (SQLSTATE 23505)"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t, s)
			}

			jobID, err := s.CreateExhaustiveSearchJob(context.Background(), test.job)

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
