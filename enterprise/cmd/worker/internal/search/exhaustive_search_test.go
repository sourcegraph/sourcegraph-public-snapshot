package search_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExhaustiveSearchHandler_Handle(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	bs := basestore.NewWithHandle(db.Handle())

	s := store.New(db, &observation.TestContext)

	userID, err := createUser(bs, "alice")
	require.NoError(t, err)

	repoID, err := createRepo(db, "repo-test")
	require.NoError(t, err)

	ctx := context.Background()
	searchJobID, err := s.CreateExhaustiveSearchJob(ctx, types.ExhaustiveSearchJob{InitiatorID: userID, Query: "test"})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupUsers(bs)
		cleanupRepos(bs)
		cleanupSearchJobs(bs)
	})

	tests := []struct {
		name        string
		mockFunc    func(*mockSearcher, *mockSearchQuery)
		record      *types.ExhaustiveSearchJob
		expectedErr error
		assertFunc  func(*testing.T)
	}{
		{
			name: "Handle job",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(searchQuery, nil).
					Once()
				searchQuery.
					On("RepositoryRevSpecs", mock.Anything).
					Return(
						[]service.RepositoryRevSpec{
							{Repository: repoID, RevisionSpecifier: "foo"},
							{Repository: repoID, RevisionSpecifier: "bar"},
						},
						nil,
					).
					Once()
			},
			record:      &types.ExhaustiveSearchJob{ID: searchJobID, InitiatorID: userID, Query: "test"},
			expectedErr: nil,
			assertFunc: func(t *testing.T) {
				jobs, err := s.ListExhaustiveSearchRepoJobs(ctx, store.ListExhaustiveSearchRepoJobsOpts{SearchJobID: searchJobID})
				require.NoError(t, err)
				require.Len(t, jobs, 2)

				assert.Equal(t, types.JobStateQueued, jobs[0].State)
				assert.Equal(t, repoID, jobs[0].RepoID)
				assert.Equal(t, "foo", jobs[0].RefSpec)

				assert.Equal(t, types.JobStateQueued, jobs[1].State)
				assert.Equal(t, repoID, jobs[1].RepoID)
				assert.Equal(t, "bar", jobs[1].RefSpec)
			},
		},
		{
			name: "Failed to create search",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(nil, errors.New("failed")).
					Once()
			},
			record:      &types.ExhaustiveSearchJob{ID: searchJobID, InitiatorID: userID, Query: "test"},
			expectedErr: errors.New("failed to create new search: failed"),
		},
		{
			name: "Failed to get repository revision specifications",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(searchQuery, nil).
					Once()
				searchQuery.
					On("RepositoryRevSpecs", mock.Anything).
					Return(
						nil,
						errors.New("failed"),
					).
					Once()
			},
			record:      &types.ExhaustiveSearchJob{ID: searchJobID, InitiatorID: userID, Query: "test"},
			expectedErr: errors.New("failed to get repository revision specifications: failed"),
		},
		{
			name: "Failed to create repo job",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(searchQuery, nil).
					Once()
				searchQuery.
					On("RepositoryRevSpecs", mock.Anything).
					Return(
						[]service.RepositoryRevSpec{
							{Repository: repoID, RevisionSpecifier: "foo"},
							{Repository: repoID, RevisionSpecifier: "bar"},
						},
						nil,
					).
					Once()
			},
			record:      &types.ExhaustiveSearchJob{InitiatorID: userID, Query: "test"},
			expectedErr: errors.New("failed to create exhaustive search repo job for repository 1: missing search job ID"),
			assertFunc: func(t *testing.T) {
				jobs, err := s.ListExhaustiveSearchRepoJobs(ctx, store.ListExhaustiveSearchRepoJobsOpts{SearchJobID: searchJobID})
				require.NoError(t, err)
				require.Len(t, jobs, 0)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			searcher := new(mockSearcher)
			searchQuery := new(mockSearchQuery)
			if test.mockFunc != nil {
				test.mockFunc(searcher, searchQuery)
			}

			handler := search.ExhaustiveSearchHandler{Store: s, Searcher: searcher}
			err := handler.Handle(ctx, logger, test.record)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if test.assertFunc != nil {
				test.assertFunc(t)
			}

			searchQuery.AssertExpectations(t)
			searcher.AssertExpectations(t)
		})
	}
}
