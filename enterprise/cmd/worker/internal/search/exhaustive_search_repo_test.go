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

func TestExhaustiveSearchRepoHandler_Handle(t *testing.T) {
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
	require.NotZero(t, searchJobID)

	repoJobID, err := s.CreateExhaustiveSearchRepoJob(ctx, types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"})
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupUsers(bs)
		cleanupRepos(bs)
		cleanupSearchJobs(bs)
		cleanupRepoJobs(bs)
	})

	tests := []struct {
		name        string
		mockFunc    func(*mockSearcher, *mockSearchQuery)
		record      *types.ExhaustiveSearchRepoJob
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
					On(
						"ResolveRepositoryRevSpec",
						mock.Anything,
						service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"},
					).
					Return(
						[]service.RepositoryRevision{
							{RepositoryRevSpec: service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"}, Revision: "foo"},
							{RepositoryRevSpec: service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"}, Revision: "foo1"},
						},
						nil,
					).
					Once()
			},
			record:      &types.ExhaustiveSearchRepoJob{ID: repoJobID, SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
			expectedErr: nil,
			assertFunc: func(t *testing.T) {
				jobs, err := s.ListExhaustiveSearchRepoRevisionJobs(ctx, store.ListExhaustiveSearchRepoRevisionJobsOpts{SearchRepoJobID: repoJobID})
				require.NoError(t, err)
				require.Len(t, jobs, 2)

				assert.Equal(t, types.JobStateQueued, jobs[0].State)
				assert.Equal(t, "foo", jobs[0].Revision)

				assert.Equal(t, types.JobStateQueued, jobs[1].State)
				assert.Equal(t, "foo1", jobs[1].Revision)
			},
		},
		{
			name:        "Failed to get search job",
			record:      &types.ExhaustiveSearchRepoJob{ID: repoJobID, SearchJobID: 999, RepoID: repoID, RefSpec: "foo"},
			expectedErr: errors.New("getting exhaustive search job: sql: no rows in result set"),
		},
		{
			name: "Failed to create search",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(nil, errors.New("failed")).
					Once()
			},
			record:      &types.ExhaustiveSearchRepoJob{ID: repoJobID, SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
			expectedErr: errors.New("creating search: failed"),
		},
		{
			name: "Failed to resolve repository revision specifications",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(searchQuery, nil).
					Once()
				searchQuery.
					On(
						"ResolveRepositoryRevSpec",
						mock.Anything,
						service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"},
					).
					Return(
						nil,
						errors.New("failed"),
					).
					Once()
			},
			record:      &types.ExhaustiveSearchRepoJob{ID: repoJobID, SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
			expectedErr: errors.New("resolving repository rev spec: failed"),
		},
		{
			name: "Failed to create revision job",
			mockFunc: func(searcher *mockSearcher, searchQuery *mockSearchQuery) {
				searcher.
					On("NewSearch", mock.Anything, "test").
					Return(searchQuery, nil).
					Once()
				searchQuery.
					On(
						"ResolveRepositoryRevSpec",
						mock.Anything,
						service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"},
					).
					Return(
						[]service.RepositoryRevision{
							{RepositoryRevSpec: service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"}, Revision: "foo"},
							{RepositoryRevSpec: service.RepositoryRevSpec{Repository: repoID, RevisionSpecifier: "foo"}, Revision: "foo1"},
						},
						nil,
					).
					Once()
			},
			record:      &types.ExhaustiveSearchRepoJob{SearchJobID: searchJobID, RepoID: repoID, RefSpec: "foo"},
			expectedErr: errors.New("creating exhaustive search repo revision job for revision \"foo\": missing search repo job ID"),
			assertFunc: func(t *testing.T) {
				jobs, err := s.ListExhaustiveSearchRepoRevisionJobs(ctx, store.ListExhaustiveSearchRepoRevisionJobsOpts{SearchRepoJobID: repoJobID})
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

			handler := search.ExhaustiveSearchRepoHandler{Store: s, Searcher: searcher}
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
