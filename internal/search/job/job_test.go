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
	test := func(input string, parser func(string) (query.Q, error)) string {
		q, _ := parser(input)
		args := &Args{
			SearchInputs: &run.SearchInputs{
				Query:        q,
				UserSettings: &schema.Settings{},
				PatternType:  query.SearchTypeLiteral,
			},
			OnSourcegraphDotCom: true,
			PatternType:         query.SearchTypeLiteral,
			Protocol:            search.Batch,
		}

		j, _ := ToSearchJob(args, q)
		return j.Name()
	}

	// Job generation for global vs non-global search
	autogold.Want("user search context", "ParallelJob{RepoSubsetText, Repo, ComputeExcludedRepos}").Equal(t, test(`foo context:@userA`, query.ParseLiteral))
	autogold.Want("universal (AKA global) search context", "ParallelJob{RepoUniverseText, Repo, ComputeExcludedRepos}").Equal(t, test(`foo context:global`, query.ParseLiteral))
	autogold.Want("universal (AKA global) search", "ParallelJob{RepoUniverseText, Repo, ComputeExcludedRepos}").Equal(t, test(`foo`, query.ParseLiteral))
	autogold.Want("nonglobal repo", "ParallelJob{RepoSubsetText, Repo, ComputeExcludedRepos}").Equal(t, test(`foo repo:sourcegraph/sourcegraph`, query.ParseLiteral))
	autogold.Want("nonglobal repo contains", "ParallelJob{RepoSubsetText, Repo, ComputeExcludedRepos}").Equal(t, test(`foo repo:contains(bar)`, query.ParseLiteral))

	// Job generation support for implied `type:repo` queries.
	autogold.Want("supported Repo job", "ParallelJob{RepoUniverseText, Repo, ComputeExcludedRepos}").Equal(t, test("ok ok", query.ParseRegexp))
	autogold.Want("supportedRepo job literal", "ParallelJob{RepoUniverseText, Repo, ComputeExcludedRepos}").Equal(t, test("ok @thing", query.ParseLiteral))
	autogold.Want("unsupported Repo job prefix", "ParallelJob{RepoUniverseText, ComputeExcludedRepos}").Equal(t, test("@nope", query.ParseRegexp))
	autogold.Want("unsupported Repo job regexp", "ParallelJob{RepoUniverseText, ComputeExcludedRepos}").Equal(t, test("foo @bar", query.ParseRegexp))

	// Job generation for other types of search
	autogold.Want("symbol", "ParallelJob{RepoUniverseSymbol, ComputeExcludedRepos}").Equal(t, test("type:symbol test", query.ParseRegexp))
	autogold.Want("commit", "ParallelJob{Commit, ComputeExcludedRepos}").Equal(t, test("type:commit test", query.ParseRegexp))
	autogold.Want("diff", "ParallelJob{Diff, ComputeExcludedRepos}").Equal(t, test("type:diff test", query.ParseRegexp))
	autogold.Want("file or commit", "JobWithOptional{Required: ParallelJob{RepoUniverseText, ComputeExcludedRepos}, Optional: Commit}").Equal(t, test("type:file type:commit test", query.ParseRegexp))
	autogold.Want("many types", "JobWithOptional{Required: ParallelJob{RepoSubsetText, Repo, ComputeExcludedRepos}, Optional: ParallelJob{RepoSubsetSymbol, Commit}}").Equal(t, test("type:file type:path type:repo type:commit type:symbol repo:test test", query.ParseRegexp))
}
