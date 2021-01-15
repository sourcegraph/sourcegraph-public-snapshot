package graphqlbackend

import (
	"flag"
	"fmt"
	"testing"

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
			SearchResultsCommon: SearchResultsCommon{
				Repos:  reposMap(timedout100...),
				Status: timedout100Status,
			},
		},
		"all": {
			SearchResults: []SearchResultResolver{mkFileMatch(&types.RepoName{Name: "found-1"}, "dir/file", 123)},
			SearchResultsCommon: SearchResultsCommon{
				IsLimitHit: true,
				Repos:      reposMap(mkRepos("found-1", "missing-1", "missing-2", "cloning-1", "timedout-1")...),
				Status: mkStatusMap(map[string]search.RepoStatus{
					"missing-1":  search.RepoStatusMissing,
					"missing-2":  search.RepoStatusMissing,
					"cloning-1":  search.RepoStatusCloning,
					"timedout-1": search.RepoStatusTimedout,
				}),
				ExcludedForks:    5,
				ExcludedArchived: 1,
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
