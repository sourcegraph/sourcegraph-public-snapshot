package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
)

const pathMatchLimit = 50

var mockSearchPathsInRepo func(ctx context.Context, repoRevs repositoryRevisions, matcher matcher) (matches []*fileMatch, limitHit bool, err error)

func searchPathsInRepo(ctx context.Context, repoRevs repositoryRevisions, matcher matcher) (matches []*fileMatch, limitHit bool, err error) {
	if mockSearchPathsInRepo != nil {
		return mockSearchPathsInRepo(ctx, repoRevs, matcher)
	}

	results, err := searchTreeForRepo(ctx, matcher, repoRevs, pathMatchLimit+1, false)
	if err != nil {
		return nil, false, err
	}
	limitHit = len(results) > pathMatchLimit

	matches = make([]*fileMatch, len(results))
	for i, res := range results {
		matches[i] = &fileMatch{JPath: res.result.(*fileResolver).path}
	}

	var workspace string
	if len(repoRevs.revs) > 0 {
		workspace = "git://" + repoRevs.repo.URI + "?" + repoRevs.revs[0].revspec + "#"
	} else {
		workspace = "git://" + repoRevs.repo.URI + "#"
	}
	for _, fm := range matches {
		fm.uri = workspace + fm.JPath
	}

	return matches, limitHit, err
}

var mockSearchPathsInRepos func(repos []*repositoryRevisions, query searchquery.Query) ([]*searchResult, *searchResultsCommon, error)

// searchPathsInRepos searches paths within a set of repositories.
func searchPathsInRepos(ctx context.Context, repos []*repositoryRevisions, query searchquery.Query) ([]*searchResult, *searchResultsCommon, error) {
	if mockSearchPathsInRepos != nil {
		return mockSearchPathsInRepos(repos, query)
	}

	// Use file and default terms as path match terms.
	valuesAsStrings := func(field string) (values, negatedValues []string) {
		for _, v := range query.Fields[field] {
			var s string
			switch {
			case v.String != nil:
				s = *v.String
			case v.Regexp != nil:
				s = v.Regexp.String()
			default:
				continue
			}
			if v.Not() {
				negatedValues = append(negatedValues, s)
			} else {
				values = append(values, s)
			}
		}
		return
	}
	includePatterns, excludePatterns := query.RegexpPatterns(searchquery.FieldFile)
	includePatterns2, excludePatterns2 := valuesAsStrings(searchquery.FieldDefault)
	includePatterns = append(includePatterns, includePatterns2...)
	excludePatterns = append(excludePatterns, excludePatterns2...)
	excludePattern := unionRegExps(excludePatterns)
	pathOptions := pathmatch.CompileOptions{
		RegExp:        true,
		CaseSensitive: query.IsCaseSensitive(),
	}
	matchPath, err := pathmatch.CompilePathPatterns(includePatterns, excludePattern, pathOptions)
	if err != nil {
		return nil, nil, err
	}

	var scorerQuery string
	if len(includePatterns) > 0 {
		// Try to extract the text-only (non-regexp) part of the query to
		// pass to stringscore, which doesn't use regexps. This is best-effort.
		scorerQuery = strings.TrimSuffix(strings.TrimPrefix(includePatterns[0], "^"), "$")
	}
	matcher := matcher{
		match:       matchPath.MatchPath,
		scorerQuery: scorerQuery,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		unflattened [][]*fileMatch
		common      = &searchResultsCommon{}
	)
	for _, repoRev := range repos {
		if len(repoRev.revs) >= 2 {
			return nil, nil, errMultipleRevsNotSupported
		}

		wg.Add(1)
		go func(repoRev repositoryRevisions) {
			defer wg.Done()
			matches, repoLimitHit, searchErr := searchPathsInRepo(ctx, repoRev, matcher)
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, searchErr); fatalErr != nil {
				if ctx.Err() != nil {
					// Our request has been canceled, we can just ignore
					// searchPathsInRepo for this repo. We only check this condition
					// here since handleRepoSearchResult handles deadlines
					// exceeded differently to canceled.
					return
				}
				err = errors.Wrapf(searchErr, "failed to search %s", repoRev.String())
				cancel()
			}
			if len(matches) > 0 {
				sort.Slice(matches, func(i, j int) bool {
					a, b := matches[i].uri, matches[j].uri
					return a > b
				})
				unflattened = append(unflattened, matches)
			}
		}(*repoRev)
	}
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	flattened := flattenFileMatches(unflattened, pathMatchLimit)
	return fileMatchesToSearchResults(flattened), common, nil
}
