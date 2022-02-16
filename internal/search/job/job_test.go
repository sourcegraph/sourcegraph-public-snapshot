package job

import (
	"testing"

	"github.com/hexops/autogold"

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
				UserSettings: &schema.Settings{},
				PatternType:  query.SearchTypeLiteral,
				Protocol:     protocol,
			},
			OnSourcegraphDotCom: true,
		}

		j, _ := ToSearchJob(args, q)
		return "\n" + PrettySexp(j) + "\n"
	}

	// Job generation for global vs non-global search
	autogold.Want("user search context", `
(PARALLEL
  RepoSubsetText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test(`foo context:@userA`, search.Streaming, query.ParseLiteral))

	autogold.Want("universal (AKA global) search context", `
(PARALLEL
  RepoUniverseText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test(`foo context:global`, search.Streaming, query.ParseLiteral))

	autogold.Want("universal (AKA global) search", `
(PARALLEL
  RepoUniverseText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test(`foo`, search.Streaming, query.ParseLiteral))

	autogold.Want("nonglobal repo", `
(PARALLEL
  RepoSubsetText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test(`foo repo:sourcegraph/sourcegraph`, search.Streaming, query.ParseLiteral))

	autogold.Want("nonglobal repo contains", `
(PARALLEL
  RepoSubsetText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test(`foo repo:contains(bar)`, search.Streaming, query.ParseLiteral))

	// Job generation support for implied `type:repo` queries.
	autogold.Want("supported Repo job", `
(PARALLEL
  RepoUniverseText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test("ok ok", search.Streaming, query.ParseRegexp))

	autogold.Want("supportedRepo job literal", `
(PARALLEL
  RepoUniverseText
  Repo
  ComputeExcludedRepos)
`).Equal(t, test("ok @thing", search.Streaming, query.ParseLiteral))

	autogold.Want("unsupported Repo job prefix", `
(PARALLEL
  RepoUniverseText
  ComputeExcludedRepos)
`).Equal(t, test("@nope", search.Streaming, query.ParseRegexp))

	autogold.Want("unsupported Repo job regexp", `
(PARALLEL
  RepoUniverseText
  ComputeExcludedRepos)
`).Equal(t, test("foo @bar", search.Streaming, query.ParseRegexp))

	// Job generation for other types of search
	autogold.Want("symbol", `
(PARALLEL
  RepoUniverseSymbol
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
  RepoUniverseText
  Commit
  ComputeExcludedRepos)
`).Equal(t, test("type:file type:commit test", search.Streaming, query.ParseRegexp))

	autogold.Want("Streaming: many types", `
(PARALLEL
  RepoSubsetText
  RepoSubsetSymbol
  Commit
  Repo
  ComputeExcludedRepos)
`).Equal(t, test("type:file type:path type:repo type:commit type:symbol repo:test test", search.Streaming, query.ParseRegexp))

	// Priority jobs for Batched search.
	autogold.Want("Batched: file or commit", `
(PRIORITY
  (REQUIRED
    (PARALLEL
      RepoUniverseText
      ComputeExcludedRepos))
  (OPTIONAL
    Commit))
`).Equal(t, test("type:file type:commit test", search.Batch, query.ParseRegexp))

	autogold.Want("Batched: many types", `
(PRIORITY
  (REQUIRED
    (PARALLEL
      RepoSubsetText
      Repo
      ComputeExcludedRepos))
  (OPTIONAL
    (PARALLEL
      RepoSubsetSymbol
      Commit)))
`).Equal(t, test("type:file type:path type:repo type:commit type:symbol repo:test test", search.Batch, query.ParseRegexp))
}
