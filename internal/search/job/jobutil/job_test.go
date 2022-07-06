package jobutil

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewPlanJob(t *testing.T) {
	cases := []struct {
		query      string
		protocol   search.Protocol
		searchType query.SearchType
		want       autogold.Value
	}{{
		query:      `foo context:@userA`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("user search context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.searchContextSpec . @userA)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"foo")
                (type . text)
                (fileMatchLimit . 500)
                (select . ))))
          (REPOPAGER
            (repoOpts.searchContextSpec . @userA)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (patternInfo . TextPatternInfo{"foo",re,filematchlimit:500})
                (numRepos . 0)
                (indexed . false)
                (useFullDeadline . true)))))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.searchContextSpec . @userA))
        (PARALLEL
          NoopJob
          (REPOSEARCH
            (repoOpts.repoFilters.0 . foo)(repoOpts.searchContextSpec . @userA)
            (filePatternsReposMustInclude . [])
            (filePatternsReposMustExclude . [])
            (contentBasedLangFilters . false)
            (mode . None)))))))`),
	}, {
		query:      `foo context:global`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("global search explicit context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"foo")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          (repoOpts.searchContextSpec . global))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.searchContextSpec . global))
        (REPOSEARCH
          (repoOpts.repoFilters.0 . foo)(repoOpts.searchContextSpec . global)
          (filePatternsReposMustInclude . [])
          (filePatternsReposMustExclude . [])
          (contentBasedLangFilters . false)
          (mode . SkipUnindexed))))))`),
	}, {
		query:      `foo`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("global search implicit context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"foo")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . foo)
          (filePatternsReposMustInclude . [])
          (filePatternsReposMustExclude . [])
          (contentBasedLangFilters . false)
          (mode . SkipUnindexed))))))`),
	}, {
		query:      `foo repo:sourcegraph/sourcegraph`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("nonglobal repo", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"foo")
                (type . text)
                (fileMatchLimit . 500)
                (select . ))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (patternInfo . TextPatternInfo{"foo",re,filematchlimit:500})
                (numRepos . 0)
                (indexed . false)
                (useFullDeadline . true)))))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . sourcegraph/sourcegraph))
        (PARALLEL
          NoopJob
          (REPOSEARCH
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)(repoOpts.repoFilters.1 . foo)
            (filePatternsReposMustInclude . [])
            (filePatternsReposMustExclude . [])
            (contentBasedLangFilters . false)
            (mode . None)))))))`),
	}, {
		query:      `ok ok`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("supported repo job", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . regex:"ok(?-s:.)*?ok")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . (?:ok).*?(?:ok))
          (filePatternsReposMustInclude . [])
          (filePatternsReposMustExclude . [])
          (contentBasedLangFilters . false)
          (mode . SkipUnindexed))))))`),
	}, {
		query:      `ok @thing`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("supported repo job literal", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"ok @thing")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . ok )
          (filePatternsReposMustInclude . [])
          (filePatternsReposMustExclude . [])
          (contentBasedLangFilters . false)
          (mode . SkipUnindexed))))))`),
	}, {
		query:      `@nope`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("unsupported repo job literal", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"@nope")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `foo @bar`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("unsupported repo job regexp", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . regex:"foo(?-s:.)*?@bar")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:symbol test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("symbol", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALSYMBOLSEARCH
          (query . sym:substr:"test")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . symbol)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . false)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:diff test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("diff", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (DIFFSEARCH
          (query . *protocol.DiffMatches(test))
          (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . true)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("streaming file or commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . content_substr:"test")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . false)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("streaming many types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"test")
                (type . text)
                (fileMatchLimit . 500)
                (select . ))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (patternInfo . TextPatternInfo{"test",re,filematchlimit:500})
                (numRepos . 0)
                (indexed . false)
                (useFullDeadline . true)))))
        (REPOPAGER
          (repoOpts.repoFilters.0 . test)
          (useIndex . yes)
          (containsRefGlobs . false)
          (PARTIALREPOS
            (ZOEKTSYMBOLSEARCH
              (query . sym:substr:"test")
              (fileMatchLimit . 500)
              (select . ))))
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts . RepoFilters: ["test"]
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . false)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . test))
        (PARALLEL
          NoopJob
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERSYMBOLSEARCH
                (patternInfo . TextPatternInfo{"test",re,filematchlimit:500})
                (numRepos . 0)
                (limit . 500))))
          (REPOSEARCH
            (repoOpts.repoFilters.0 . test)(repoOpts.repoFilters.1 . test)
            (filePatternsReposMustInclude . [])
            (filePatternsReposMustExclude . [])
            (contentBasedLangFilters . false)
            (mode . None)))))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("batched file or commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . content_substr:"test")
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . text)
          (fileMatchLimit . 500)
          (select . )
          )
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . false)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("batched many types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"test")
                (type . text)
                (fileMatchLimit . 500)
                (select . ))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (patternInfo . TextPatternInfo{"test",re,filematchlimit:500})
                (numRepos . 0)
                (indexed . false)
                (useFullDeadline . true)))))
        (REPOPAGER
          (repoOpts.repoFilters.0 . test)
          (useIndex . yes)
          (containsRefGlobs . false)
          (PARTIALREPOS
            (ZOEKTSYMBOLSEARCH
              (query . sym:substr:"test")
              (fileMatchLimit . 500)
              (select . ))))
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts . RepoFilters: ["test"]
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
          (diff . false)
          (limit . 500)
          (includeModifiedFiles . false))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . test))
        (PARALLEL
          NoopJob
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (useIndex . yes)
            (containsRefGlobs . false)
            (PARTIALREPOS
              (SEARCHERSYMBOLSEARCH
                (patternInfo . TextPatternInfo{"test",re,filematchlimit:500})
                (numRepos . 0)
                (limit . 500))))
          (REPOSEARCH
            (repoOpts.repoFilters.0 . test)(repoOpts.repoFilters.1 . test)
            (filePatternsReposMustInclude . [])
            (filePatternsReposMustExclude . [])
            (contentBasedLangFilters . false)
            (mode . None)))))))`),
	}, {
		query:      `(type:commit or type:diff) (a or b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		// TODO this output doesn't look right. There shouldn't be any zoekt or repo jobs
		want: autogold.Want("complex commit diff", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (OR
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (COMMITSEARCH
            (query . (*protocol.MessageMatches((?:a)|(?:b))))
            (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
            (diff . false)
            (limit . 500)
            (includeModifiedFiles . false))
          (REPOSCOMPUTEEXCLUDED
            )
          (OR
            NoopJob
            NoopJob))))
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (DIFFSEARCH
            (query . (*protocol.DiffMatches((?:a)|(?:b))))
            (repoOpts . RepoFilters: []
MinusRepoFilters: []
CommitAfter:
Visibility: Any
NoForks: true
OnlyCloned: true
NoArchived: true
)
            (diff . true)
            (limit . 500)
            (includeModifiedFiles . false))
          (REPOSCOMPUTEEXCLUDED
            )
          (OR
            NoopJob
            NoopJob))))))`),
	}, {
		query:      `(type:repo a) or (type:file b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("disjunct types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (OR
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            )
          (REPOSEARCH
            (repoOpts.repoFilters.0 . a)
            (filePatternsReposMustInclude . [])
            (filePatternsReposMustExclude . [])
            (contentBasedLangFilters . false)
            (mode . None)))))
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . content_substr:"b")
            (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
            (includePrivate . true)
            (type . text)
            (fileMatchLimit . 500)
            (select . )
            )
          (REPOSCOMPUTEEXCLUDED
            )
          NoopJob)))))`),
	}, {
		query:      `type:symbol a or b`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("symbol with or", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (onSourcegraphDotCom . true)
  (protocol . Streaming)
  (features . )
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALSYMBOLSEARCH
          (query . (or sym:substr:"a" sym:substr:"b"))
          (repoScope . ["(and branch=\"HEAD\" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived)"])
          (includePrivate . true)
          (type . symbol)
          (fileMatchLimit . 500)
          (select . )
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (OR
          NoopJob
          NoopJob)))))`),
	}}

	for _, tc := range cases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			plan, err := query.Pipeline(query.Init(tc.query, tc.searchType))
			require.NoError(t, err)

			inputs := &run.SearchInputs{
				UserSettings:        &schema.Settings{},
				PatternType:         query.SearchTypeLiteral,
				Protocol:            tc.protocol,
				OnSourcegraphDotCom: true,
			}

			j, err := NewPlanJob(inputs, plan)
			require.NoError(t, err)

			tc.want.Equal(t, "\n"+printer.SexpPretty(j))
		})
	}
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
		j, _ := toFlatJobs(inputs, b)
		return "\n" + printer.SexpPretty(j) + "\n"
	}

	autogold.Want("root limit for streaming search", `
(REPOSEARCH
  (repoOpts.repoFilters.0 . foo)
  (filePatternsReposMustInclude . [])
  (filePatternsReposMustExclude . [])
  (contentBasedLangFilters . false)
  (mode . SkipUnindexed))
`).Equal(t, test("foo", search.Streaming))

	autogold.Want("root limit for batch search", `
(REPOSEARCH
  (repoOpts.repoFilters.0 . foo)
  (filePatternsReposMustInclude . [])
  (filePatternsReposMustExclude . [])
  (contentBasedLangFilters . false)
  (mode . SkipUnindexed))
`).Equal(t, test("foo", search.Batch))
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
	}, {
		input:  `repo:contains.file(Dockerfile)`,
		output: autogold.Want("108", `{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":["Dockerfile"],"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repohasfile:Dockerfile`,
		output: autogold.Want("109", `{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","FilePatternsReposMustInclude":["Dockerfile"],"FilePatternsReposMustExclude":null,"PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
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
