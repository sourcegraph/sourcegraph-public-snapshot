package graphqlbackend

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var updateGolden = flag.Bool("update", false, "Update testdata goldens")

func TestSearchProgress(t *testing.T) {
	cases := map[string]*SearchResultsResolver{
		"empty": {},
		"all": {
			SearchResults: []SearchResultResolver{mkFileMatch(&types.Repo{Name: "found-1"}, "dir/file", 123)},
			searchResultsCommon: searchResultsCommon{
				limitHit: true,
				repos:    mkRepos("found-1", "missing-1", "missing-2", "cloning-1", "timedout-1"),
				missing:  mkRepos("missing-1", "missing-2"),
				cloning:  mkRepos("cloning-1"),
				timedout: mkRepos("timedout-1"),
				excluded: excludedRepos{
					forks:    5,
					archived: 1,
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
