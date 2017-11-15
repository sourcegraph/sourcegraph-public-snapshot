package graphqlbackend

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

type searchResults2 struct {
	results  []*fileMatch
	limitHit bool
	cloning  []string
	missing  []string
	alert    *searchAlert
}

func (sr *searchResults2) Results() []*fileMatch {
	return sr.results
}

func (sr *searchResults2) LimitHit() bool {
	return sr.limitHit
}

func (sr *searchResults2) Cloning() []string {
	if sr.cloning == nil {
		return []string{}
	}
	return sr.cloning
}

func (sr *searchResults2) Missing() []string {
	if sr.missing == nil {
		return []string{}
	}
	return sr.missing
}

func (r searchResults2) Alert() *searchAlert { return r.alert }

func (r *searchResolver2) Results(ctx context.Context) (*searchResults2, error) {
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

	if len(patternsToCombine) == 0 {
		return nil, errors.New("no query terms or regexp specified")
	}

	args := repoSearchArgs{
		Query: &patternInfo{
			IsRegExp:                     true,
			IsCaseSensitive:              r.combinedQuery.isCaseSensitive(),
			FileMatchLimit:               300,
			Pattern:                      strings.Join(patternsToCombine, ".*?"), // "?" makes it prefer shorter matches
			IncludePatterns:              r.combinedQuery.fieldValues[searchFieldFile].Values(),
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: r.combinedQuery.isCaseSensitive(),
		},
		Repositories: repos,
	}
	if excludePatterns := r.combinedQuery.fieldValues[minusField(searchFieldFile)].Values(); len(excludePatterns) > 0 {
		pat := unionRegExps(excludePatterns)
		args.Query.ExcludePattern = &pat
	}

	results, err := r.root.SearchRepos(ctx, &args)
	if results != nil {
		if len(missingRepoRevs) > 0 {
			results.alert = r.alertForMissingRepoRevs(missingRepoRevs)
		}
	}
	return results, err
}
