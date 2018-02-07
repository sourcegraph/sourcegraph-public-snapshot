package graphqlbackend

import (
	"context"
	"regexp"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
)

var mockSearchRepositories func(args *repoSearchArgs) ([]*searchResult, *searchResultsCommon, error)

// searchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name ("URI") must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func searchRepositories(ctx context.Context, args *repoSearchArgs, query searchquery.Query) (res []*searchResult, common *searchResultsCommon, err error) {
	if mockSearchRepositories != nil {
		return mockSearchRepositories(args)
	}

	// Only proceed if the query consists solely of type:, repo:, repogroup:, and default (no field) tokens.
	// Matching repositories based whether they contain files at a certain path (etc.) is not yet implemented.
	for field := range query.Fields {
		if field != searchquery.FieldRepo && field != searchquery.FieldRepoGroup && field != searchquery.FieldType && field != searchquery.FieldDefault {
			return nil, nil, nil
		}
	}

	pattern, err := regexp.Compile(args.query.Pattern)
	if err != nil {
		return nil, nil, err
	}

	common = &searchResultsCommon{}
	var results []*searchResult
	for _, repo := range args.repos {
		if pattern.MatchString(string(repo.repo.URI)) {
			results = append(results, &searchResult{repo: &repositoryResolver{repo: repo.repo}})
		}
	}
	return results, common, nil
}
