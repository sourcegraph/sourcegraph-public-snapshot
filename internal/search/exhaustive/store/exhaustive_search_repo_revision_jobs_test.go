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

func TestStore_CreateExhaustiveSearchRepoRevisionJob(t *testing.T) {
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

	ctx := context.Background()
	searchJobID, err := s.CreateExhaustiveSearchJob(
		ctx,
		types.ExhaustiveSearchJob{InitiatorID: userID, Query: "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchRepoRevisionJob"},
	)
	require.NoError(t, err)

	repoJobID, err := s.CreateExhaustiveSearchRepoJob(
		ctx,
		types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "main"},
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		job         types.ExhaustiveSearchRepoRevisionJob
		expectedErr error
	}{
		{
			name: "New job",
			job: types.ExhaustiveSearchRepoRevisionJob{
				SearchRepoJobID: repoJobID,
				Revision:        "main",
			},
			expectedErr: nil,
		},
		{
			name: "Missing revision",
			job: types.ExhaustiveSearchRepoRevisionJob{
				SearchRepoJobID: repoJobID,
			},
			expectedErr: errors.New("missing revision"),
		},
		{
			name: "Missing repo job ID",
			job: types.ExhaustiveSearchRepoRevisionJob{
				Revision: "main",
			},
			expectedErr: errors.New("missing search repo job ID"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				cleanupRevJobs(bs)
			})

			jobID, err := s.CreateExhaustiveSearchRepoRevisionJob(ctx, test.job)

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

func TestStore_ListExhaustiveSearchRepoRevisionJobs(t *testing.T) {
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

	ctx := context.Background()
	searchJobID, err := s.CreateExhaustiveSearchJob(
		ctx,
		types.ExhaustiveSearchJob{InitiatorID: userID, Query: "repo:^github\\.com/hashicorp/errwrap$ CreateExhaustiveSearchRepoJob"},
	)
	require.NoError(t, err)

	repoJobID, err := s.CreateExhaustiveSearchRepoJob(
		ctx,
		types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
	)
	require.NoError(t, err)

	_, err = s.CreateExhaustiveSearchRepoRevisionJob(
		ctx,
		types.ExhaustiveSearchRepoRevisionJob{SearchRepoJobID: repoJobID, Revision: "foo"},
	)
	require.NoError(t, err)
	_, err = s.CreateExhaustiveSearchRepoRevisionJob(
		ctx,
		types.ExhaustiveSearchRepoRevisionJob{SearchRepoJobID: repoJobID, Revision: "bar"},
	)
	require.NoError(t, err)

	first := 1
	firstJobID := 1

	tests := []struct {
		name           string
		opts           store.ListExhaustiveSearchRepoRevisionJobsOpts
		expectedLength int
		expectedErr    error
	}{
		{
			name: "All",
			opts: store.ListExhaustiveSearchRepoRevisionJobsOpts{
				SearchRepoJobID: searchJobID,
			},
			expectedLength: 2,
			expectedErr:    nil,
		},
		{
			name: "First",
			opts: store.ListExhaustiveSearchRepoRevisionJobsOpts{
				First:           &first,
				SearchRepoJobID: searchJobID,
			},
			expectedLength: 1,
			expectedErr:    nil,
		},
		{
			name: "Last",
			opts: store.ListExhaustiveSearchRepoRevisionJobsOpts{
				After:           &firstJobID,
				SearchRepoJobID: searchJobID,
			},
			expectedLength: 1,
			expectedErr:    nil,
		},
		{
			name:        "Missing job ID",
			opts:        store.ListExhaustiveSearchRepoRevisionJobsOpts{},
			expectedErr: errors.New("missing search repo job ID"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jobs, err := s.ListExhaustiveSearchRepoRevisionJobs(ctx, test.opts)

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
