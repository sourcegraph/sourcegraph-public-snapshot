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
		Name      string
		Query     string
		WantPager autogold.Value
		WantJob   autogold.Value
	}{
		{
			Name:  "glob",
			Query: `type:file index:no repo:foo rev:*refs/heads/dev* content`,
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . true)
  (repoOpts.repoFilters . [foo@*refs/heads/dev*])
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"content",re,nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"content",re,nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:  "",
			Query: "type:file index:no content",
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"content",re,nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"content",re,nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
		{
			Name:  "",
			Query: "type:file index:no foo.*bar patterntype:regexp",
			WantPager: autogold.Expect(`
(REPOPAGER
  (containsRefGlobs . false)
  (repoOpts.useIndex . no)
  (PARTIALREPOS
    (SEARCHERTEXTSEARCH
      (useFullDeadline . true)
      (patternInfo . TextPatternInfo{"foo\\.\\*bar",re,nopath,filematchlimit:1000000})
      (numRepos . 0)
      (pathRegexps . [])
      (indexed . false))))
`),
			WantJob: autogold.Expect(`
(SEARCHERTEXTSEARCH
  (useFullDeadline . true)
  (patternInfo . TextPatternInfo{"foo\\.\\*bar",re,nopath,filematchlimit:1000000})
  (numRepos . 1)
  (pathRegexps . [])
  (indexed . false))
`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			searchType := query.SearchTypeLiteral
			plan, err := query.Pipeline(query.Init(tc.Query, searchType))
			require.NoError(t, err)

			inputs := &search.Inputs{
				Plan:         plan,
				Query:        plan.ToQ(),
				UserSettings: &schema.Settings{},
				PatternType:  searchType,
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

// The queries are validated before the reach exhaustive search, hence we only
// have to worry about valid queries we don't want to process for now.
func TestNewExhaustive_negative(t *testing.T) {
	defaultInputs := &search.Inputs{
		UserSettings: &schema.Settings{},
		PatternType:  query.SearchTypeLiteral,
		Protocol:     search.Exhaustive,
		Features:     &search.Features{},
	}

	tc := []struct {
		query  string
		inputs *search.Inputs
	}{
		// >1 type filter.
		{
			query:  `type:file index:no type:diff content`,
			inputs: defaultInputs,
		},
		{
			query:  `type:file index:no type:path content`,
			inputs: defaultInputs,
		},
		// AND, OR
		{
			query:  `type:file index:no repo:repo1 rev:branch1 content1 OR content2`,
			inputs: defaultInputs,
		},
		{
			query:  `type:file index:no repo:repo1 rev:branch1 content1 AND content2`,
			inputs: defaultInputs,
		},
		{
			query:  `type:file index:no (repo:repo1 or repo:repo2) content`,
			inputs: defaultInputs,
		},
		// catch-all regex
		{
			query: `type:file index:no r:.* .*`,
			inputs: &search.Inputs{
				UserSettings: &schema.Settings{},
				PatternType:  query.SearchTypeRegex,
				Protocol:     search.Exhaustive,
				Features:     &search.Features{},
			},
		},
		// predicates
		{
			query:  `type:file index:no repohasfile:foo.bar content`,
			inputs: defaultInputs,
		},
		{
			query:  `type:file index:no file:has.content("content")`,
			inputs: defaultInputs,
		},
		{
			query:  `type:file index:no repo:has.path("src") content`,
			inputs: defaultInputs,
		},
	}

	for _, c := range tc {
		t.Run(c.query, func(t *testing.T) {
			plan, err := query.Pipeline(query.Init(c.query, c.inputs.PatternType))
			require.NoError(t, err)

			c.inputs.Plan = plan
			c.inputs.Query = plan.ToQ()
			c.inputs.OriginalQuery = c.query

			_, err = NewExhaustive(c.inputs)
			require.Error(t, err)
		})
	}
}
