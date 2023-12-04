package repos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func TestComputeExcludedJob(t *testing.T) {
	tests := []struct {
		name           string
		repoFilters    []string
		numArchived    int
		numForks       int
		expectedResult streaming.SearchEvent
	}{
		{
			name:        "compute excluded repos",
			repoFilters: []string{"sourcegraph/.*"},
			numArchived: 3,
			numForks:    42,
			expectedResult: streaming.SearchEvent{
				Stats: streaming.Stats{
					ExcludedArchived: 3,
					ExcludedForks:    42,
				},
			},
		},
		{
			name:        "compute excluded repos with single matching repo",
			repoFilters: []string{"^gitlab\\.com/sourcegraph/sourcegraph$"},
			numArchived: 10,
			numForks:    2,
			expectedResult: streaming.SearchEvent{
				Stats: streaming.Stats{
					ExcludedArchived: 0,
					ExcludedForks:    0,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsedFilters := make([]query.ParsedRepoFilter, len(tc.repoFilters))
			for i, repoFilter := range tc.repoFilters {
				parsedFilter, err := query.ParseRepositoryRevisions(repoFilter)
				if err != nil {
					t.Fatalf("unexpected error parsing repo filter %s", repoFilter)
				}
				parsedFilters[i] = parsedFilter
			}

			repoStore := dbmocks.NewMockRepoStore()
			repoStore.CountFunc.SetDefaultHook(func(_ context.Context, opt database.ReposListOptions) (int, error) {
				// Verify that the include patterns passed to the DB match the repo filters
				numFilters := len(parsedFilters)
				require.Equal(t, numFilters, len(opt.IncludePatterns))

				for i, repo := range opt.IncludePatterns {
					require.Equal(t, parsedFilters[i].Repo, repo)
				}

				if opt.OnlyForks {
					return tc.numForks, nil
				} else if opt.OnlyArchived {
					return tc.numArchived, nil
				} else {
					return numFilters, nil
				}
			})

			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefaultReturn(repoStore)

			var result streaming.SearchEvent
			streamCollector := streaming.StreamFunc(func(event streaming.SearchEvent) {
				result = event
			})

			j := ComputeExcludedJob{RepoOpts: search.RepoOptions{RepoFilters: parsedFilters}}
			alert, err := j.Run(context.Background(), job.RuntimeClients{DB: db}, streamCollector)
			require.Nil(t, alert)
			require.NoError(t, err)
			require.Equal(t, tc.expectedResult, result)
		})
	}
}
