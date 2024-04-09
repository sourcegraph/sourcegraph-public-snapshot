package jobutil

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewExhaustive(t *testing.T) {
	// When reading the autogold expectations things we currently want to see:
	//
	//  - Simple jobs that are essentially just REPOPAGER and SEARCHERTEXTSEARCH
	//  - Only use searcher, not zoekt (indexed . false)
	//  - Avoid normal timeouts (useFullDeadline . true)
	//  - High limits (filematchlimit:1000000)
	repoRevs := &search.RepositoryRevisions{
		Repo: types.MinimalRepo{
			ID:   1,
			Name: "repo",
		},
		Revs: []string{"dev1"},
	}

	cases := []struct {
		Name       string
		Query      string
		WantPager  autogold.Value
		WantJob    autogold.Value
		SearchType query.SearchType
	}{
		{
			Name:       "case sensitive lang match",
			Query:      `type:file index:no lang:cpp content case:yes`,
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"content",case,nopath,filematchlimit:1000000,lang:cpp,F:"(?i)(?:\\.cpp$)|(?:\\.c\\+\\+$)|(?:\\.cc$)|(?:\\.cp$)|(?:\\.cxx$)|(?:\\.h$)|(?:\\.h\\+\\+$)|(?:\\.hh$)|(?:\\.hpp$)|(?:\\.hxx$)|(?:\\.inc$)|(?:\\.inl$)|(?:\\.ino$)|(?:\\.ipp$)|(?:\\.ixx$)|(?:\\.re$)|(?:\\.tcc$)|(?:\\.tpp$)"})
      (numRepos . 0)
      (pathRegexps . [(?i)(?:\.cpp$)|(?:\.c\+\+$)|(?:\.cc$)|(?:\.cp$)|(?:\.cxx$)|(?:\.h$)|(?:\.h\+\+$)|(?:\.hh$)|(?:\.hpp$)|(?:\.hxx$)|(?:\.inc$)|(?:\.inl$)|(?:\.ino$)|(?:\.ipp$)|(?:\.ixx$)|(?:\.re$)|(?:\.tcc$)|(?:\.tpp$)])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"content",case,nopath,filematchlimit:1000000,lang:cpp,F:"(?i)(?:\\.cpp$)|(?:\\.c\\+\\+$)|(?:\\.cc$)|(?:\\.cp$)|(?:\\.cxx$)|(?:\\.h$)|(?:\\.h\\+\\+$)|(?:\\.hh$)|(?:\\.hpp$)|(?:\\.hxx$)|(?:\\.inc$)|(?:\\.inl$)|(?:\\.ino$)|(?:\\.ipp$)|(?:\\.ixx$)|(?:\\.re$)|(?:\\.tcc$)|(?:\\.tpp$)"})
  (numRepos . 1)
  (pathRegexps . [(?i)(?:\.cpp$)|(?:\.c\+\+$)|(?:\.cc$)|(?:\.cp$)|(?:\.cxx$)|(?:\.h$)|(?:\.h\+\+$)|(?:\.hh$)|(?:\.hpp$)|(?:\.hxx$)|(?:\.inc$)|(?:\.inl$)|(?:\.ino$)|(?:\.ipp$)|(?:\.ixx$)|(?:\.re$)|(?:\.tcc$)|(?:\.tpp$)])
  (indexed . false))
`),
		},
		{
			Name:       "glob",
			Query:      `type:file index:no repo:foo rev:*refs/heads/dev* content`,
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . true)
  (repoOpts.repoFilters . [foo@*refs/heads/dev*])
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"content",nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"content",nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "only pattern",
			Query:      "type:file index:no content",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"content",nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"content",nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "keyword search",
			Query:      "type:file index:no foo bar baz patterntype:keyword",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"foo bar baz",nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"foo bar baz",nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "boolean query",
			Query:      "type:file index:no (foo OR bar) AND baz patterntype:keyword",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{(("foo" OR "bar") AND "baz"),nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{(("foo" OR "bar") AND "baz"),nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "regexp",
			Query:      "type:file index:no foo.*bar patterntype:regexp",
			SearchType: query.SearchTypeRegex,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{/foo.*bar/,nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{/foo.*bar/,nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "repo:has.file predicate",
			Query:      "type:file index:no foo repo:has.file(go.mod)",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (repoOpts.hasFileContent[0].path . go.mod)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"foo",nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"foo",nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "repo:has.meta predicate",
			Query:      "type:file index:no foo repo:has.meta(cognition)",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (repoOpts.hasKVPs[0].key . cognition)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"foo",nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"foo",nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:       "type:diff author:alice",
			Query:      "index:no type:diff author:alice",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (repoOpts.onlyCloned . true)
  (PARTIALREPOS
    (DIFFSEARCH
      (includeModifiedFiles . false)
      (query . *protocol.AuthorMatches(alice))
      (diff . true)
      (limit . 1000000))))
`),
			WantJob: autogold.Expect(`
(DIFFSEARCH
  (includeModifiedFiles . false)
  (query . *protocol.AuthorMatches(alice))
  (diff . true)
  (limit . 1000000))
`),
		},
		{
			Name:       "type:commit author:alice",
			Query:      "index:no type:commit author:alice",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (repoOpts.onlyCloned . true)
  (PARTIALREPOS
    (COMMITSEARCH
      (includeModifiedFiles . false)
      (query . *protocol.AuthorMatches(alice))
      (diff . false)
      (limit . 1000000))))
`),
			WantJob: autogold.Expect(`
(COMMITSEARCH
  (includeModifiedFiles . false)
  (query . *protocol.AuthorMatches(alice))
  (diff . false)
  (limit . 1000000))
`),
		},
		{
			Name:       "type:path f:search.go",
			Query:      "index:no type:path f:search.go",
			SearchType: query.SearchTypeLiteral,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{//,nocontent,filematchlimit:1000000,f:"search.go"})
      (numRepos . 0)
      (pathRegexps . [(?i)search.go])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{//,nocontent,filematchlimit:1000000,f:"search.go"})
  (numRepos . 1)
  (pathRegexps . [(?i)search.go])
  (indexed . false))
`),
		},
		{
			Name:       "patternType:structural if ... else ...",
			Query:      "index:no patterntype:structural if ... else ...",
			SearchType: query.SearchTypeStructural,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (STRUCTURALSEARCH
      (useFullDeadline . true)
      (useIndex . no)
      (patternInfo.query . "if :[_] else :[_]")
      (patternInfo.isStructural . true)
      (patternInfo.fileMatchLimit . 1000000)
      (patternInfo.index . no))))
`),
			WantJob: autogold.Expect(`
(STRUCTURALSEARCH
  (useFullDeadline . true)
  (useIndex . no)
  (patternInfo.query . "if :[_] else :[_]")
  (patternInfo.isStructural . true)
  (patternInfo.fileMatchLimit . 1000000)
  (patternInfo.index . no))
`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			plan, err := query.Pipeline(query.Init(tc.Query, tc.SearchType))
			require.NoError(t, err)

			inputs := &search.Inputs{
				Plan:         plan,
				Query:        plan.ToQ(),
				UserSettings: &schema.Settings{},
				PatternType:  tc.SearchType,
				Protocol:     search.Exhaustive,
				Features:     &search.Features{},
			}

			exhaustive, err := NewExhaustive(inputs)
			require.NoError(t, err)

			tc.WantPager.Equal(t, sPrintSexpMax(exhaustive.repoPagerJob))

			tc.WantJob.Equal(t, sPrintSexpMax(exhaustive.Job(repoRevs)))
		})
	}
}

func sPrintSexpMax(j job.Describer) string {
	return "\n" + printer.SexpVerbose(j, job.VerbosityMax, true) + "\n"
}

// The queries are validated before they reach exhaustive search, hence we only
// have to worry about valid queries we don't want to process for now.
func TestNewExhaustive_negative(t *testing.T) {
	tc := []struct {
		query      string
		searchType query.SearchType
	}{
		// multiple jobs needed
		{query: `type:file index:no (repo:repo1 or repo:repo2) content`, searchType: query.SearchTypeLiteral},
		// catch-all regex
		{query: `type:file index:no r:.* .*`, searchType: query.SearchTypeRegex},
		{query: `type:file index:no r:repo .*`, searchType: query.SearchTypeRegex},
		// file predicates
		{query: `type:file index:no file:has.content(content)`, searchType: query.SearchTypeLiteral},
		{query: `type:file index:no file:has.owner(owner)`, searchType: query.SearchTypeLiteral},
		{query: `type:file index:no file:contains.content(content)`, searchType: query.SearchTypeLiteral},
		{query: `type:file index:no file:has.contributor(contributor)`, searchType: query.SearchTypeLiteral},
		// unsupported types
		{query: `index:no type:repo`, searchType: query.SearchTypeLiteral},
		{query: `index:no type:symbol`, searchType: query.SearchTypeLiteral},
		{query: `index:no foo select:file.owners`, searchType: query.SearchTypeLiteral},
		// structural search with AND/OR
		{query: `index:no patterntype:structural if ... OR switch ... `, searchType: query.SearchTypeStructural},
		{query: `index:no patterntype:structural if ... AND switch ... `, searchType: query.SearchTypeStructural},
	}

	for _, c := range tc {
		t.Run("", func(t *testing.T) {
			plan, err := query.Pipeline(query.Init(c.query, c.searchType))
			require.NoError(t, err)

			inputs := &search.Inputs{
				Plan:         plan,
				Query:        plan.ToQ(),
				UserSettings: &schema.Settings{},
				PatternType:  c.searchType,
				Protocol:     search.Exhaustive,
				Features:     &search.Features{},
			}

			_, err = NewExhaustive(inputs)
			require.Error(t, err, "failed query: %q", c.query)
		})
	}
}
