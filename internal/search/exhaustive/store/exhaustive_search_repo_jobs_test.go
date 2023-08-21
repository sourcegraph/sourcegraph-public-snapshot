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

func TestStore_CreateExhaustiveSearchRepoJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bs := basestore.NewWithHandle(db.Handle())

	t.Cleanup(func() {
		cleanupUsers(bs)
		cleanupRepos(bs)
		cleanupSearchJobs(bs)
	})

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)
	repoID, err := createRepo(db, "repo-test")
	require.NoError(t, err)

	s := store.New(db, &observation.TestContext)

	searchJobID, err := s.CreateExhaustiveSearchJob(
		context.Background(),
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
			t.Cleanup(func() {
				cleanupRepoJobs(bs)
			})

			jobID, err := s.CreateExhaustiveSearchRepoJob(context.Background(), test.job)

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

func TestStore_ListExhaustiveSearchRepoJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bs := basestore.NewWithHandle(db.Handle())

	t.Cleanup(func() {
		cleanupUsers(bs)
		cleanupRepos(bs)
		cleanupSearchJobs(bs)
		cleanupRepoJobs(bs)
	})

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)
	repoID, err := createRepo(db, "repo-test")
	require.NoError(t, err)

	s := store.New(db, &observation.TestContext)

	searchJobID, err := s.CreateExhaustiveSearchJob(
		context.Background(),
		types.ExhaustiveSearchJob{InitiatorID: userID, Query: "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchRepoJob"},
	)
	require.NoError(t, err)

	_, err = s.CreateExhaustiveSearchRepoJob(
		context.Background(),
		types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
	)
	require.NoError(t, err)
	_, err = s.CreateExhaustiveSearchRepoJob(
		context.Background(),
		types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "bar"},
	)
	require.NoError(t, err)

	first := 1
	firstJobID := 1

	tests := []struct {
		name           string
		opts           store.ListExhaustiveSearchRepoJobsOpts
		expectedLength int
		expectedErr    error
	}{
		{
			name: "All",
			opts: store.ListExhaustiveSearchRepoJobsOpts{
				SearchJobID: searchJobID,
			},
			expectedLength: 2,
			expectedErr:    nil,
		},
		{
			name: "First",
			opts: store.ListExhaustiveSearchRepoJobsOpts{
				First:       &first,
				SearchJobID: searchJobID,
			},
			expectedLength: 1,
			expectedErr:    nil,
		},
		{
			name: "Last",
			opts: store.ListExhaustiveSearchRepoJobsOpts{
				After:       &firstJobID,
				SearchJobID: searchJobID,
			},
			expectedLength: 1,
			expectedErr:    nil,
		},
		{
			name:        "Missing search job ID",
			opts:        store.ListExhaustiveSearchRepoJobsOpts{},
			expectedErr: errors.New("missing search job ID"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jobs, err := s.ListExhaustiveSearchRepoJobs(context.Background(), test.opts)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(jobs), test.expectedLength)
			}
		})
	}
}
