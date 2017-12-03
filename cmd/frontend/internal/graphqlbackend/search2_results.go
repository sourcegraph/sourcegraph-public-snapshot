package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// searchResultsCommon contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type searchResultsCommon struct {
	limitHit bool     // whether the limit on results was hit
	cloning  []string // repos that could not be searched because they were still being cloned
	missing  []string // repos that could not be searched because they do not exist

	// timedout usually contains repos that haven't finished being fetched yet.
	// This should only happen for large repos and the searcher caches are
	// purged.
	timedout []string
}

func (c *searchResultsCommon) LimitHit() bool {
	return c.limitHit
}

func (c *searchResultsCommon) Cloning() []string {
	if c.cloning == nil {
		return []string{}
	}
	return c.cloning
}

func (c *searchResultsCommon) Missing() []string {
	if c.missing == nil {
		return []string{}
	}
	return c.missing
}

func (c *searchResultsCommon) Timedout() []string {
	if c.timedout == nil {
		return []string{}
	}
	return c.timedout
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *searchResultsCommon) update(other searchResultsCommon) {
	c.limitHit = c.limitHit || other.limitHit

	appendUnique := func(dst *[]string, src []string) {
		dstSet := make(map[string]struct{}, len(*dst))
		for _, s := range *dst {
			dstSet[s] = struct{}{}
		}
		for _, s := range src {
			if _, present := dstSet[s]; !present {
				*dst = append(*dst, s)
			}
		}
	}
	appendUnique(&c.cloning, other.cloning)
	appendUnique(&c.missing, other.missing)
	appendUnique(&c.timedout, other.timedout)
}

type searchResults2 struct {
	results []*searchResult
	searchResultsCommon
	alert *searchAlert
}

func (sr *searchResults2) Results() []*searchResult {
	return sr.results
}

func (sr *searchResults2) ResultCount() int32 {
	return int32(len(sr.results))
}

func (sr *searchResults2) ApproximateResultCount() string {
	if sr.alert != nil {
		return "?"
	}
	if sr.limitHit || len(sr.missing) > 0 || len(sr.cloning) > 0 {
		return fmt.Sprintf("%d+", len(sr.results))
	}
	return strconv.Itoa(len(sr.results))
}

func (sr *searchResults2) Alert() *searchAlert { return sr.alert }

func (r *searchResolver2) Results(ctx context.Context) (*searchResults2, error) {
	return r.doResults(ctx, "")
}

func (r *searchResolver2) doResults(ctx context.Context, forceOnlyResultType string) (*searchResults2, error) {
	repos, missingRepoRevs, _, overLimit, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		alert, err := r.alertForNoResolvedRepos(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResults2{alert: alert}, nil
	}
	if overLimit {
		alert, err := r.alertForOverRepoLimit(ctx)
		if err != nil {
			return nil, err
		}
		return &searchResults2{alert: alert}, nil
	}

	var patternsToCombine []string
	for _, term := range r.combinedQuery.fieldValues[""] {
		if term.Value == "" {
			continue
		}

		// Treat quoted strings as literal strings to match, not regexps.
		var value string
		if term.Quoted {
			value = regexp.QuoteMeta(term.Value)
		} else {
			value = term.Value
		}
		patternsToCombine = append(patternsToCombine, value)
	}
	args := repoSearchArgs{
		query: &patternInfo{
			IsRegExp:                     true,
			IsCaseSensitive:              r.combinedQuery.isCaseSensitive(),
			FileMatchLimit:               300,
			Pattern:                      strings.Join(patternsToCombine, ".*?"), // "?" makes it prefer shorter matches
			IncludePatterns:              r.combinedQuery.fieldValues[searchFieldFile].Values(),
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: r.combinedQuery.isCaseSensitive(),
		},
		repos: repos,
	}
	if excludePatterns := r.combinedQuery.fieldValues[minusField(searchFieldFile)].Values(); len(excludePatterns) > 0 {
		pat := unionRegExps(excludePatterns)
		args.query.ExcludePattern = &pat
	}

	// Determine which types of results to return.
	var searchFuncs []func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error)
	var resultTypes []string
	if forceOnlyResultType != "" {
		resultTypes = []string{forceOnlyResultType}
	} else {
		resultTypes = r.combinedQuery.fieldValues[searchFieldType].Values()
		if len(resultTypes) == 0 {
			resultTypes = []string{"file"} // TODO(sqs)
		}
	}
	seenResultTypes := make(map[string]struct{}, len(resultTypes))
	for _, resultType := range resultTypes {
		if _, seen := seenResultTypes[resultType]; seen {
			continue
		}
		seenResultTypes[resultType] = struct{}{}
		switch resultType {
		case "file":
			if len(patternsToCombine) == 0 {
				return nil, errors.New("no query terms or regexp specified")
			}
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchRepos(ctx, &args)
			})
		case "diff":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitDiffsInRepos(ctx, &args, r.combinedQuery)
			})
		case "commit":
			searchFuncs = append(searchFuncs, func(ctx context.Context) ([]*searchResult, *searchResultsCommon, error) {
				return searchCommitLogInRepos(ctx, &args, r.combinedQuery)
			})
		}
	}

	// Run all search funcs.
	var results searchResults2
	for _, searchFunc := range searchFuncs {
		results1, common1, err := searchFunc(ctx)
		if err != nil {
			return nil, err
		}
		if results1 == nil && common1 == nil {
			continue
		}
		results.results = append(results.results, results1...)
		// TODO(sqs): combine diff and commit results that refer to the same underlying
		// commit (and match on the commit's diff and message, respectively).
		results.searchResultsCommon.update(*common1)
	}

	if len(missingRepoRevs) > 0 {
		results.alert = r.alertForMissingRepoRevs(missingRepoRevs)
	}

	return &results, nil
}

type searchResult struct {
	fileMatch *fileMatch
	diff      *commitSearchResult
}

func (g *searchResult) ToFileMatch() (*fileMatch, bool) { return g.fileMatch, g.fileMatch != nil }
func (g *searchResult) ToCommitSearchResult() (*commitSearchResult, bool) {
	return g.diff, g.diff != nil
}
