package graphqlbackend

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
)

var mockSearchRepositories func(args *search.Args) ([]*searchResultResolver, *searchResultsCommon, error)

// searchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name ("URI") must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func searchRepositories(ctx context.Context, args *search.Args, limit int32) (res []*searchResultResolver, common *searchResultsCommon, err error) {
	if mockSearchRepositories != nil {
		return mockSearchRepositories(args)
	}

	fieldWhitelist := map[string]struct{}{
		query.FieldRepo:      struct{}{},
		query.FieldRepoGroup: struct{}{},
		query.FieldType:      struct{}{},
		query.FieldDefault:   struct{}{},
		query.FieldIndex:     struct{}{},
		query.FieldCount:     struct{}{},
		query.FieldMax:       struct{}{},
		query.FieldTimeout:   struct{}{},
		query.FieldFork:      struct{}{},
		query.FieldArchived:  struct{}{},
	}
	// Don't return repo results if the search contains fields that aren't on the whitelist.
	// Matching repositories based whether they contain files at a certain path (etc.) is not yet implemented.
	for field := range args.Query.Fields {
		if _, ok := fieldWhitelist[field]; !ok {
			return nil, nil, nil
		}
	}

	pattern, err := regexp.Compile(args.Pattern.Pattern)
	if err != nil {
		return nil, nil, err
	}

	common = &searchResultsCommon{}
	var results []*searchResultResolver
	for _, repo := range args.Repos {
		if len(results) == int(limit) {
			common.limitHit = true
			break
		}
		if pattern.MatchString(string(repo.Repo.URI)) {
			results = append(results, &searchResultResolver{repo: &repositoryResolver{repo: repo.Repo}})
		}
	}
	return results, common, nil
}
