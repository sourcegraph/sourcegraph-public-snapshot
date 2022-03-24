package job

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestToSearchInputs(t *testing.T) {
	test := func(input string, protocol search.Protocol, parser func(string) (query.Q, error)) string {
		q, _ := parser(input)
		args := &Args{
			SearchInputs: &run.SearchInputs{
				UserSettings:        &schema.Settings{},
				PatternType:         query.SearchTypeLiteral,
				Protocol:            protocol,
				OnSourcegraphDotCom: true,
			},
		}

		j, _ := ToSearchJob(args, q, database.NewMockDB())
		return "\n" + PrettySexp(j) + "\n"
	}

	// Job generation for global vs non-global search
	autogold.Want("user search context", `
(PARALLEL
  REPOPAGER
    (PARALLEL
      ZoektRepoSubset
      Searcher))
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test(`foo context:@userA`, search.Streaming, query.ParseLiteral))

	autogold.Want("universal (AKA global) search context", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test(`foo context:global`, search.Streaming, query.ParseLiteral))

	autogold.Want("universal (AKA global) search", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test(`foo`, search.Streaming, query.ParseLiteral))

	autogold.Want("nonglobal repo", `
(PARALLEL
  REPOPAGER
    (PARALLEL
      ZoektRepoSubset
      Searcher))
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test(`foo repo:sourcegraph/sourcegraph`, search.Streaming, query.ParseLiteral))

	autogold.Want("nonglobal repo contains", `
(PARALLEL
  REPOPAGER
    (PARALLEL
      ZoektRepoSubset
      Searcher))
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test(`foo repo:contains(bar)`, search.Streaming, query.ParseLiteral))

	// Job generation support for implied `type:repo` queries.
	autogold.Want("supported Repo job", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test("ok ok", search.Streaming, query.ParseRegexp))

	autogold.Want("supportedRepo job literal", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test("ok @thing", search.Streaming, query.ParseLiteral))

	autogold.Want("unsupported Repo job prefix", `
(PARALLEL
  ZoektGlobalSearch
  ComputeExcludedRepos)
`).Equal(t, test("@nope", search.Streaming, query.ParseRegexp))

	autogold.Want("unsupported Repo job regexp", `
(PARALLEL
  ZoektGlobalSearch
  ComputeExcludedRepos)
`).Equal(t, test("foo @bar", search.Streaming, query.ParseRegexp))

	// Job generation for other types of search
	autogold.Want("symbol", `
(PARALLEL
  RepoUniverseSymbolSearch
  ComputeExcludedRepos)
`).Equal(t, test("type:symbol test", search.Streaming, query.ParseRegexp))

	autogold.Want("commit", `
(PARALLEL
  Commit
  ComputeExcludedRepos)
`).Equal(t, test("type:commit test", search.Streaming, query.ParseRegexp))

	autogold.Want("diff", `
(PARALLEL
  Diff
  ComputeExcludedRepos)
`).Equal(t, test("type:diff test", search.Streaming, query.ParseRegexp))

	autogold.Want("Streaming: file or commit", `
(PARALLEL
  ZoektGlobalSearch
  Commit
  ComputeExcludedRepos)
`).Equal(t, test("type:file type:commit test", search.Streaming, query.ParseRegexp))

	autogold.Want("Streaming: many types", `
(PARALLEL
  REPOPAGER
    (PARALLEL
      ZoektRepoSubset
      Searcher))
  REPOPAGER
    (PARALLEL
      ZoektSymbolSearch
      SymbolSearcher))
  Commit
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test("type:file type:path type:repo type:commit type:symbol repo:test test", search.Streaming, query.ParseRegexp))

	// Priority jobs for Batched search.
	autogold.Want("Batched: file or commit", `
(PRIORITY
  (REQUIRED
    (PARALLEL
      ZoektGlobalSearch
      ComputeExcludedRepos))
  (OPTIONAL
    Commit))
`).Equal(t, test("type:file type:commit test", search.Batch, query.ParseRegexp))

	autogold.Want("Batched: many types", `
(PRIORITY
  (REQUIRED
    (PARALLEL
      REPOPAGER
        (PARALLEL
          ZoektRepoSubset
          Searcher))
      RepoSearch
      ComputeExcludedRepos))
  (OPTIONAL
    (PARALLEL
      REPOPAGER
        (PARALLEL
          ZoektSymbolSearch
          SymbolSearcher))
      Commit)))
`).Equal(t, test("type:file type:path type:repo type:commit type:symbol repo:test test", search.Batch, query.ParseRegexp))
}

func TestToEvaluateJob(t *testing.T) {
	test := func(input string, protocol search.Protocol) string {
		q, _ := query.ParseLiteral(input)
		args := &Args{
			SearchInputs: &run.SearchInputs{
				UserSettings:        &schema.Settings{},
				PatternType:         query.SearchTypeLiteral,
				Protocol:            protocol,
				OnSourcegraphDotCom: true,
			},
		}

		b, _ := query.ToBasicQuery(q)
		j, _ := ToEvaluateJob(args, b, database.NewMockDB())
		return "\n" + PrettySexp(j) + "\n"
	}

	autogold.Want("root limit for streaming search", `
(TIMEOUT
  20s
  (LIMIT
    500
    (PARALLEL
      ZoektGlobalSearch
      RepoSearch
      ComputeExcludedRepos)))
`).Equal(t, test("foo", search.Streaming))

	autogold.Want("root limit for batch search", `
(TIMEOUT
  20s
  (LIMIT
    30
    (PARALLEL
      ZoektGlobalSearch
      RepoSearch
      ComputeExcludedRepos)))
`).Equal(t, test("foo", search.Batch))
}
