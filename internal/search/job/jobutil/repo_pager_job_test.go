package jobutil

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_setRepos(t *testing.T) {
	// Static test data
	indexed := &zoekt.IndexedRepoRevs{
		RepoRevs: map[api.RepoID]*search.RepositoryRevisions{
			1: {Repo: types.MinimalRepo{Name: "indexed"}},
		},
	}
	unindexed := []*search.RepositoryRevisions{
		{Repo: types.MinimalRepo{Name: "unindexed"}},
	}

	// Test function
	test := func(job job.Job) string {
		job = setRepos(job, indexed, unindexed)
		return "\n" + PrettyJSONVerbose(job)
	}

	autogold.Want("set repos for jobs that need them", `
{
  "PARALLEL": [
    {
      "ZoektRepoSubsetTextSearchJob": {
        "Repos": {
          "RepoRevs": {
            "1": {
              "Repo": {
                "ID": 0,
                "Name": "indexed",
                "Stars": 0
              },
              "Revs": null
            }
          }
        },
        "Query": null,
        "Typ": "",
        "FileMatchLimit": 0,
        "Select": null
      }
    },
    {
      "SearcherTextSearchJob": {
        "PatternInfo": null,
        "Repos": [
          {
            "Repo": {
              "ID": 0,
              "Name": "unindexed",
              "Stars": 0
            },
            "Revs": null
          }
        ],
        "Indexed": false,
        "UseFullDeadline": false,
        "Features": {
          "ContentBasedLangFilters": false,
          "HybridSearch": false
        }
      }
    }
  ]
}`).Equal(t, test(
		NewParallelJob(
			&zoekt.RepoSubsetTextSearchJob{},
			&searcher.TextSearchJob{},
		),
	))
}
