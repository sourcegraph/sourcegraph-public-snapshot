package jobutil

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSexp(t *testing.T) {
	autogold.Want("simple sexp", "(TIMEOUT 50ms (AND NoopJob NoopJob))").Equal(t, Sexp(
		NewTimeoutJob(
			50*1_000_000,
			NewAndJob(
				NewNoopJob(),
				NewNoopJob()))))

	autogold.Want("pretty sexp exhaustive cases", `
(FILTER
  SubRepoPermissions
  (LIMIT
    100
    (TIMEOUT
      50ms
      (PARALLEL
        (AND
          NoopJob
          NoopJob)
        (OR
          NoopJob
          NoopJob)
        (AND
          NoopJob
          NoopJob)))))
`).Equal(t, fmt.Sprintf("\n%s\n", PrettySexp(
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
}

func TestPrettyMermaid(t *testing.T) {
	autogold.Want("simple mermaid", `
flowchart TB
0([AND])
  0---1
  1([NoopJob])
  0---2
  2([NoopJob])
  `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
		NewAndJob(
			NewNoopJob(),
			NewNoopJob()))))

	autogold.Want("big mermaid", `
flowchart TB
0([FILTER])
  0---1
  1[SubRepoPermissions]
  0---2
  2([LIMIT])
    2---3
    3[100]
    2---4
    4([TIMEOUT])
      4---5
      5[50ms]
      4---6
      6([PARALLEL])
        6---7
        7([AND])
          7---8
          8([NoopJob])
          7---9
          9([NoopJob])
          6---10
        10([OR])
          10---11
          11([NoopJob])
          10---12
          12([NoopJob])
          6---13
        13([AND])
          13---14
          14([NoopJob])
          13---15
          15([NoopJob])
          `).Equal(t, fmt.Sprintf("\n%s", PrettyMermaid(
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
}

func TestPrettyJSON(t *testing.T) {
	autogold.Want("big JSON summary", `
{
  "FILTER": {
    "LIMIT": {
      "TIMEOUT": {
        "PARALLEL": [
          {
            "AND": [
              "NoopJob",
              "NoopJob"
            ]
          },
          {
            "OR": [
              "NoopJob",
              "NoopJob"
            ]
          },
          {
            "AND": [
              "NoopJob",
              "NoopJob"
            ]
          }
        ]
      },
      "value": "50ms"
    },
    "value": 100
  },
  "value": "SubRepoPermissions"
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
		return PrettyJSONVerbose(j)
	}

	autogold.Want("full fidelity JSON output", `
{
  "TIMEOUT": {
    "LIMIT": {
      "PARALLEL": [
        {
          "REPOPAGER": {
            "ZoektRepoSubsetTextSearchJob": {
              "Repos": null,
              "Query": {
                "Pattern": "bar",
                "CaseSensitive": false,
                "FileName": false,
                "Content": false
              },
              "Typ": "text",
              "FileMatchLimit": 500,
              "Select": []
            }
          }
        },
        {
          "ReposComputeExcludedJob": {
            "RepoOpts": {
              "RepoFilters": [
                "foo"
              ],
              "MinusRepoFilters": null,
              "Dependencies": null,
              "Dependents": null,
              "CaseSensitiveRepoFilters": false,
              "SearchContextSpec": "",
              "CommitAfter": "",
              "Visibility": "Any",
              "Limit": 0,
              "Cursors": null,
              "ForkSet": false,
              "NoForks": true,
              "OnlyForks": false,
              "OnlyCloned": false,
              "ArchivedSet": false,
              "NoArchived": true,
              "OnlyArchived": false
            }
          }
        },
        {
          "PARALLEL": [
            {
              "REPOPAGER": {
                "SearcherTextSearchJob": {
                  "PatternInfo": {
                    "Pattern": "bar",
                    "IsNegated": false,
                    "IsRegExp": true,
                    "IsStructuralPat": false,
                    "CombyRule": "",
                    "IsWordMatch": false,
                    "IsCaseSensitive": false,
                    "FileMatchLimit": 500,
                    "Index": "yes",
                    "Select": [],
                    "IncludePatterns": null,
                    "ExcludePattern": "",
                    "FilePatternsReposMustInclude": null,
                    "FilePatternsReposMustExclude": null,
                    "PathPatternsAreCaseSensitive": false,
                    "PatternMatchesContent": true,
                    "PatternMatchesPath": true,
                    "Languages": null
                  },
                  "Repos": null,
                  "Indexed": false,
                  "UseFullDeadline": true,
                  "Features": {
                    "ContentBasedLangFilters": false,
                    "HybridSearch": false
                  }
                }
              }
            },
            {
              "RepoSearchJob": {
                "RepoOpts": {
                  "RepoFilters": [
                    "foo",
                    "bar"
                  ],
                  "MinusRepoFilters": null,
                  "Dependencies": null,
                  "Dependents": null,
                  "CaseSensitiveRepoFilters": false,
                  "SearchContextSpec": "",
                  "CommitAfter": "",
                  "Visibility": "Any",
                  "Limit": 0,
                  "Cursors": null,
                  "ForkSet": false,
                  "NoForks": true,
                  "OnlyForks": false,
                  "OnlyCloned": false,
                  "ArchivedSet": false,
                  "NoArchived": true,
                  "OnlyArchived": false
                },
                "FilePatternsReposMustInclude": null,
                "FilePatternsReposMustExclude": null,
                "Features": {
                  "ContentBasedLangFilters": false,
                  "HybridSearch": false
                },
                "Mode": 0
              }
            }
          ]
        }
      ]
    },
    "value": 500
  },
  "value": "20s"
}`).Equal(t, fmt.Sprintf("\n%s", test("repo:foo bar")))
}
