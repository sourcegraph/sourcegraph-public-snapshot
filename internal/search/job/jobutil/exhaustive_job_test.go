package jobutil

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
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
	}{{
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
	}}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			searchType := query.SearchTypeLiteral
			plan, err := query.Pipeline(query.Init(tc.Query, searchType))
			require.NoError(t, err)

			inputs := &search.Inputs{
				Plan:         plan,
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
