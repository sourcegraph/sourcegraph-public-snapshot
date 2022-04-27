package jobutil

import (
	"encoding/json"
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

	autogold.Want("commit with and", `
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
            Commit
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            Commit
            ComputeExcludedRepos))))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        Commit
        (AND
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              ComputeExcludedRepos)))))))
`).Equal(t, test("type:commit a and b"))
	autogold.Want("commit with or", `
BASE:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (OR
        (PARALLEL
          Commit
          ComputeExcludedRepos)
        (PARALLEL
          Commit
          ComputeExcludedRepos)))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        Commit
        (OR
          (PARALLEL
            NoopJob
            ComputeExcludedRepos)
          (PARALLEL
            NoopJob
            ComputeExcludedRepos))))))
`).Equal(t, test("type:commit a or b"))
	autogold.Want("diff with and", `
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
            Diff
            ComputeExcludedRepos))
        (LIMIT
          40000
          (PARALLEL
            Diff
            ComputeExcludedRepos))))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        Diff
        (AND
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              ComputeExcludedRepos))
          (LIMIT
            40000
            (PARALLEL
              NoopJob
              ComputeExcludedRepos)))))))
`).Equal(t, test("type:diff a and b"))

	autogold.Want("diff with or", `
BASE:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (OR
        (PARALLEL
          Diff
          ComputeExcludedRepos)
        (PARALLEL
          Diff
          ComputeExcludedRepos)))))

OPTIMIZED:

(ALERT
  (TIMEOUT
    20s
    (LIMIT
      500
      (PARALLEL
        Diff
        (OR
          (PARALLEL
            NoopJob
            ComputeExcludedRepos)
          (PARALLEL
            NoopJob
            ComputeExcludedRepos))))))
`).Equal(t, test("type:diff a or b"))
}

func TestToTextPatternInfo(t *testing.T) {
	cases := []struct {
		input  string
		output autogold.Value
	}{{
		input:  `type:repo archived`,
		output: autogold.Want("01", `{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo archived archived:yes`,
		output: autogold.Want("02", `{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux`,
		output: autogold.Want("04", `{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux fork:yes`,
		output: autogold.Want("05", `{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" patterntype:regexp type:file`,
		output: autogold.Want("10", `{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" -repo:go-diff patterntype:regexp type:file`,
		output: autogold.Want("11", `{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ String case:yes type:file`,
		output: autogold.Want("12", `{"Pattern":"String","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":true,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":true,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal type:file`,
		output: autogold.Want("13", `{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal count:1 type:file`,
		output: autogold.Want("14", `{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":1,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:only patterntype:regexp type:file`,
		output: autogold.Want("15", `{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:no patterntype:regexp type:file`,
		output: autogold.Want("16", `{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ doesnot734734743734743exist`,
		output: autogold.Want("17", `{"Pattern":"doesnot734734743734743exist","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ type:commit test`,
		output: autogold.Want("21", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ type:diff main`,
		output: autogold.Want("22", `{"Pattern":"main","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ repohascommitafter:"2019-01-01" test patterntype:literal`,
		output: autogold.Want("23", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `^func.*$ patterntype:regexp index:only type:file`,
		output: autogold.Want("24", `{"Pattern":"^func.*$","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `fork:only patterntype:regexp FORK_SENTINEL`,
		output: autogold.Want("25", `{"Pattern":"FORK_SENTINEL","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `\bfunc\b lang:go type:file patterntype:regexp`,
		output: autogold.Want("26", `{"Pattern":"\\bfunc\\b","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) index:only patterntype:structural count:3`,
		output: autogold.Want("29", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) lang:go rule:'where "backcompat" == "backcompat"' patterntype:structural`,
		output: autogold.Want("30", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"where \"backcompat\" == \"backcompat\"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$@adde71 make(:[1]) index:no patterntype:structural count:3`,
		output: autogold.Want("31", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ file:^README\.md "basic :[_] access :[_]" patterntype:structural`,
		output: autogold.Want("32", `{"Pattern":"\"basic :[_] access :[_]\"","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^README\\.md"],"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `no results for { ... } raises alert repo:^github\.com/sgtest/go-diff$`,
		output: autogold.Want("34", `{"Pattern":"no results for \\{ \\.\\.\\. \\} raises alert","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ patternType:regexp \ and /`,
		output: autogold.Want("49", `{"Pattern":"(?:\\ and).*?(?:/)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ (not .svg) patterntype:literal`,
		output: autogold.Want("52", `{"Pattern":"\\.svg","IsNegated":true,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (Fetches OR file:language-server.ts)`,
		output: autogold.Want("72", `{"Pattern":"Fetches","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ ((file:^renovate\.json extends) or file:progress.ts createProgressProvider)`,
		output: autogold.Want("73", `{"Pattern":"extends","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^renovate\\.json"],"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) author:felix yarn`,
		output: autogold.Want("74", `{"Pattern":"yarn","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) subscription after:"june 11 2019" before:"june 13 2019"`,
		output: autogold.Want("75", `{"Pattern":"subscription","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/go-diff$@garo/lsif-indexing-campaign:test-already-exist-pr or repo:^github\.com/sgtest/sourcegraph-typescript$) file:README.md #`,
		output: autogold.Want("78", `{"Pattern":"#","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["README.md"],"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/sourcegraph-typescript$ or repo:^github\.com/sgtest/go-diff$) package diff provides`,
		output: autogold.Want("79", `{"Pattern":"package diff provides","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains(file:noexist.go) test`,
		output: autogold.Want("83", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains(file:go.mod) count:100 fmt`,
		output: autogold.Want("87", `{"Pattern":"fmt","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":100,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `type:commit LSIF`,
		output: autogold.Want("90", `{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:contains(file:diff.pb.go) type:commit LSIF`,
		output: autogold.Want("91", `{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:repo`,
		output: autogold.Want("93", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["repo"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:file`,
		output: autogold.Want("96", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["file"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:content`,
		output: autogold.Want("98", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["content"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize`,
		output: autogold.Want("99", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:commit`,
		output: autogold.Want("100", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["commit"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:symbol`,
		output: autogold.Want("101", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal type:symbol HunkNoChunksize select:symbol`,
		output: autogold.Want("102", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `foo\d "bar*" patterntype:regexp`,
		output: autogold.Want("105", `{"Pattern":"(?:foo\\d).*?(?:bar\\*)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `patterntype:regexp // literal slash`,
		output: autogold.Want("107", `{"Pattern":"(?://).*?(?:literal).*?(?:slash)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":null,"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}}

	test := func(input string) string {
		searchType := overrideSearchType(input, query.SearchTypeLiteral)
		plan, err := query.Pipeline(query.Init(input, searchType))
		if err != nil {
			return "Error"
		}
		if len(plan) == 0 {
			return "Empty"
		}
		b := plan[0]
		types, _ := b.ToParseTree().StringValues(query.FieldType)
		mode := search.Batch
		resultTypes := computeResultTypes(types, b, query.SearchTypeLiteral)
		p := toTextPatternInfo(b, resultTypes, mode)
		v, _ := json.Marshal(p)
		return string(v)
	}

	for _, tc := range cases {
		t.Run(tc.output.Name(), func(t *testing.T) {
			tc.output.Equal(t, test(tc.input))
		})
	}
}

func overrideSearchType(input string, searchType query.SearchType) query.SearchType {
	q, err := query.Parse(input, query.SearchTypeLiteral)
	q = query.LowercaseFieldNames(q)
	if err != nil {
		// If parsing fails, return the default search type. Any actual
		// parse errors will be raised by subsequent parser calls.
		return searchType
	}
	query.VisitField(q, "patterntype", func(value string, _ bool, _ query.Annotation) {
		switch value {
		case "regex", "regexp":
			searchType = query.SearchTypeRegex
		case "literal":
			searchType = query.SearchTypeLiteral
		case "structural":
			searchType = query.SearchTypeStructural
		}
	})
	return searchType
}
