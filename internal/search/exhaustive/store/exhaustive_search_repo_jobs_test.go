package store_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestStore_CreateExhaustiveSearchRepoJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	bs := basestore.NewWithHandle(db.Handle())

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)
	repoID, err := createRepo(db, "repo-test")
	require.NoError(t, err)

	ctx := actor.WithActor(context.Background(), &actor.Actor{
		UID: userID,
	})

	s := store.New(db, &observation.TestContext)

	searchJobID, err := s.CreateExhaustiveSearchJob(
		ctx,
		types.ExhaustiveSearchJob{InitiatorID: userID, Query: "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchRepoJob"},
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		job         types.ExhaustiveSearchRepoJob
		expectedErr error
	}{
		{
			name: "New job",
			job: types.ExhaustiveSearchRepoJob{
				SearchJobID: searchJobID,
				RepoID:      repoID,
				RefSpec:     "bar:baz",
			},
			expectedErr: nil,
		},
		{
			name: "Missing repo ID",
			job: types.ExhaustiveSearchRepoJob{
				SearchJobID: searchJobID,
				RefSpec:     "bar:baz",
			},
			expectedErr: errors.New("missing repo ID"),
		},
		{
			name: "Missing search job ID",
			job: types.ExhaustiveSearchRepoJob{
				RepoID:  repoID,
				RefSpec: "bar:baz",
			},
			expectedErr: errors.New("missing search job ID"),
		},
		{
			name: "Missing ref spec",
			job: types.ExhaustiveSearchRepoJob{
				SearchJobID: searchJobID,
				RepoID:      repoID,
			},
			expectedErr: errors.New("missing ref spec"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jobID, err := s.CreateExhaustiveSearchRepoJob(ctx, test.job)

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
