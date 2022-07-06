package printer

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	. "github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPrettyJSON(t *testing.T) {
	autogold.Want("big JSON summary", `
{
  "SubRepoPermsFilterJob": {
    "LimitJob": {
      "TimeoutJob": {
        "ParallelJob": {
          "AndJob.0": {
            "NoopJob.0": {},
            "NoopJob.1": {}
          },
          "AndJob.1": {
            "NoopJob.0": {},
            "NoopJob.1": {}
          },
          "OrJob": {
            "NoopJob.0": {},
            "NoopJob.1": {}
          }
        },
        "timeout": "50ms"
      },
      "limit": 100
    }
  }
}`).Equal(t, fmt.Sprintf("\n%s", PrettyJSON(
		NewFilterJob(
			NewLimitJob(
				100,
				NewTimeoutJob(
					50*1_000_000,
					NewParallelJob(
						NewAndJob(
							NewNoopJob(),
							NewNoopJob()),
						NewOrJob(
							NewNoopJob(),
							NewNoopJob()),
						NewAndJob(
							NewNoopJob(),
							NewNoopJob()))))))))
	test := func(input string) string {
		q, _ := query.ParseLiteral(input)
		b, _ := query.ToBasicQuery(q)
		inputs := &run.SearchInputs{
			UserSettings: &schema.Settings{},
			Protocol:     search.Streaming,
		}
		j, _ := NewBasicJob(inputs, b)
		return PrettyJSONVerbose(j, job.VerbosityMax)
	}

	autogold.Want("full fidelity JSON output", `
{
  "TimeoutJob": {
    "LimitJob": {
      "ParallelJob": {
        "ParallelJob": {
          "RepoPagerJob": {
            "PartialReposJob": {
              "SearcherTextSearchJob": {
                "indexed": false,
                "numRepos": 0,
                "patternInfo": "TextPatternInfo{\"bar\",re,filematchlimit:500}",
                "useFullDeadline": true
              }
            },
            "containsRefGlobs": false,
            "repoOpts": "RepoFilters: [\"foo\"]\nMinusRepoFilters: []\nCommitAfter: \nVisibility: Any\nNoForks: true\nNoArchived: true\n",
            "useIndex": "yes"
          },
          "RepoSearchJob": {
            "contentBasedLangFilters": false,
            "filePatternsReposMustExclude": "[]",
            "filePatternsReposMustInclude": "[]",
            "mode": "None",
            "repoOpts": "RepoFilters: [\"foo\" \"bar\"]\nMinusRepoFilters: []\nCommitAfter: \nVisibility: Any\nNoForks: true\nNoArchived: true\n"
          }
        },
        "RepoPagerJob": {
          "PartialReposJob": {
            "ZoektRepoSubsetTextSearchJob": {
              "fileMatchLimit": 500,
              "query": "substr:\"bar\"",
              "select": "",
              "type": "text"
            }
          },
          "containsRefGlobs": false,
          "repoOpts": "RepoFilters: [\"foo\"]\nMinusRepoFilters: []\nCommitAfter: \nVisibility: Any\nNoForks: true\nNoArchived: true\n",
          "useIndex": "yes"
        },
        "ReposComputeExcludedJob": {
          "repoOpts": "RepoFilters: [\"foo\"]\nMinusRepoFilters: []\nCommitAfter: \nVisibility: Any\nNoForks: true\nNoArchived: true\n"
        }
      },
      "limit": 500
    },
    "timeout": "20s"
  }
}`).Equal(t, fmt.Sprintf("\n%s", test("repo:foo bar")))
}
