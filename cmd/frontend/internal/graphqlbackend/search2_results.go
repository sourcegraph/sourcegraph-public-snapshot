package graphqlbackend

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

func (r *searchResolver2) Results(ctx context.Context) (*searchResults, error) {
	repos, _, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
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
	if len(repos) == 0 {
		return nil, errors.New("no repositories included")
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

	return r.root.SearchRepos(ctx, &args)
}
