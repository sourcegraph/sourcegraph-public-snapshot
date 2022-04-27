package jobutil

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestToSearchInputs(t *testing.T) {
	test := func(input string, protocol search.Protocol, parser func(string) (query.Q, error)) string {
		q, _ := parser(input)
		b, err := query.ToBasicQuery(q)
		require.NoError(t, err)
		inputs := &run.SearchInputs{
			UserSettings:        &schema.Settings{},
			PatternType:         query.SearchTypeLiteral,
			Protocol:            protocol,
			OnSourcegraphDotCom: true,
		}

		j, _ := ToSearchJob(inputs, b)
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
		inputs := &run.SearchInputs{
			UserSettings:        &schema.Settings{},
			PatternType:         query.SearchTypeLiteral,
			Protocol:            protocol,
			OnSourcegraphDotCom: true,
		}

		b, _ := query.ToBasicQuery(q)
		j, _ := ToEvaluateJob(inputs, b)
		return "\n" + PrettySexp(j) + "\n"
	}

	autogold.Want("root limit for streaming search", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test("foo", search.Streaming))

	autogold.Want("root limit for batch search", `
(PARALLEL
  ZoektGlobalSearch
  RepoSearch
  ComputeExcludedRepos)
`).Equal(t, test("foo", search.Batch))
}

func Test_optimizeJobs(t *testing.T) {
	test := func(input string) string {
		plan, _ := query.Pipeline(query.InitLiteral(input))
		inputs := &run.SearchInputs{
			UserSettings:        &schema.Settings{},
			PatternType:         query.SearchTypeLiteral,
			Protocol:            search.Streaming,
			OnSourcegraphDotCom: true,
		}

		baseJob, _ := NewJob(inputs, plan, IdentityPass)
		optimizedJob, _ := NewJob(inputs, plan, OptimizationPass)
		return "\nBASE:\n\n" + PrettySexp(baseJob) + "\n\nOPTIMIZED:\n\n" + PrettySexp(optimizedJob) + "\n"
	}

	autogold.Want("optimize basic expression (Zoekt Text Global)", `
BASE:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (AND
        (LIMIT
          40000
          (PARALLEL
            ZoektGlobalSearch
            RepoSearch
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            ZoektGlobalSearch
            RepoSearch
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            ZoektGlobalSearch
            RepoSearch
            ComputeExcludedRepos))))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        ZoektGlobalSearch
        (AND
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              RepoSearch
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              RepoSearch
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              RepoSearch
              ComputeExcludedRepos)))))))
`).Equal(t, test("foo and bar and not baz"))

	autogold.Want("optimize repo-qualified expression (Zoekt Text over repos)", `
BASE:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (AND
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektRepoSubset
                Searcher))
            RepoSearch
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektRepoSubset
                Searcher))
            RepoSearch
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektRepoSubset
                Searcher))
            RepoSearch
            ComputeExcludedRepos))))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        REPOPAGER
          ZoektRepoSubset)
        (AND
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  Searcher))
              RepoSearch
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  Searcher))
              RepoSearch
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  Searcher))
              RepoSearch
              ComputeExcludedRepos)))))))
`).Equal(t, test("repo:derp foo and bar not baz"))

	autogold.Want("optimize qualified repo with type:symbol expression (Zoekt Symbol over repos)", `
BASE:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (AND
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektSymbolSearch
                SymbolSearcher))
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektSymbolSearch
                SymbolSearcher))
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            REPOPAGER
              (PARALLEL
                ZoektSymbolSearch
                SymbolSearcher))
            ComputeExcludedRepos))))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        REPOPAGER
          ZoektSymbolSearch)
        (AND
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  SymbolSearcher))
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  SymbolSearcher))
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              REPOPAGER
                (PARALLEL
                  NoopJob
                  SymbolSearcher))
              ComputeExcludedRepos)))))))
`).Equal(t, test("repo:derp foo and bar not baz type:symbol"))

}
