package graphqlbackend

import (
	"context"
	"regexp"
)

func (r *searchResolver2) Results(ctx context.Context) (*searchResults, error) {
	repos, _, err := r.resolveRepositories(ctx, nil)
	if err != nil {
		return nil, err
	}

	patternsToCombine := make([]string, 0, len(r.query.fieldValues[""])+len(r.query.fieldValues[searchFieldRegExp]))
	for _, term := range r.query.fieldValues[""] {
		patternsToCombine = append(patternsToCombine, regexp.QuoteMeta(term))
	}
	for _, pattern := range r.query.fieldValues[searchFieldRegExp] {
		patternsToCombine = append(patternsToCombine, pattern)
	}

	args := repoSearchArgs{
		Query: &patternInfo{
			IsRegExp:                     true,
			IsCaseSensitive:              r.query.isCaseSensitive(),
			FileMatchLimit:               300,
			Pattern:                      unionRegExps(patternsToCombine),
			IncludePatterns:              r.query.fieldValues[searchFieldFile],
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: r.query.isCaseSensitive(),
		},
		Repositories: repos,
	}
	if excludePatterns := r.query.fieldValues[minusField(searchFieldFile)]; len(excludePatterns) > 0 {
		pat := unionRegExps(excludePatterns)
		args.Query.ExcludePattern = &pat
	}

	return r.root.SearchRepos(ctx, &args)
}
