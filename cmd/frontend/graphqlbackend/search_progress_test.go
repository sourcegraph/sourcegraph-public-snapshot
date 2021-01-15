package graphqlbackend

import (
	"flag"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var updateGolden = flag.Bool("update", false, "Update testdata goldens")

func TestSearchProgress(t *testing.T) {
	var timedout100 []*types.RepoName
	var timedout100Status search.RepoStatusMap
	for i := 0; i < 100; i++ {
		r := mkRepos(fmt.Sprintf("timedout-%d", i))[0]
		timedout100 = append(timedout100, r)
		timedout100Status.Update(r.ID, search.RepoStatusTimedout)
	}

	cases := map[string]*SearchResultsResolver{
		"empty": {},
		"timedout100": {
			searchResultsCommon: searchResultsCommon{
				repos:  reposMap(timedout100...),
				status: timedout100Status,
			},
		},
		"all": {
			SearchResults: []SearchResultResolver{mkFileMatch(&types.RepoName{Name: "found-1"}, "dir/file", 123)},
			searchResultsCommon: searchResultsCommon{
				limitHit: true,
				repos:    reposMap(mkRepos("found-1", "missing-1", "missing-2", "cloning-1", "timedout-1")...),
				status: mkStatusMap(map[string]search.RepoStatus{
					"missing-1":  search.RepoStatusMissing,
					"missing-2":  search.RepoStatusMissing,
					"cloning-1":  search.RepoStatusCloning,
					"timedout-1": search.RepoStatusTimedout,
				}),
				excluded: repos.ExcludedRepos{
					Forks:    5,
					Archived: 1,
				},
			},
		},
	}

	for name, sr := range cases {
		t.Run(name, func(t *testing.T) {
			got := sr.Progress()
			got.DurationMs = 0 // clear out non-deterministic field
			testutil.AssertGolden(t, "testdata/golden/"+t.Name()+".json", *updateGolden, got)
		})
	}
}
