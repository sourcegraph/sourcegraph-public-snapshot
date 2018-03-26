package graphqlbackend

import (
	"context"
	"regexp"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
)

var mockSearchRepositories func(args *repoSearchArgs) ([]*searchResultResolver, *searchResultsCommon, error)

// searchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name ("URI") must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func searchRepositories(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int32) (res []*searchResultResolver, common *searchResultsCommon, err error) {
	if mockSearchRepositories != nil {
		return mockSearchRepositories(args)
	}

	pattern, err := regexp.Compile(args.query.Pattern)
	if err != nil {
		return nil, nil, err
	}

	common = &searchResultsCommon{}
	var results []*searchResultResolver
	for _, repo := range args.repos {
		if len(results) == int(limit) {
			common.limitHit = true
			break
		}
		if pattern.MatchString(string(repo.repo.URI)) {
			results = append(results, &searchResultResolver{repo: &repositoryResolver{repo: repo.repo}})
		}
	}
	return results, common, nil
}
