package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func (r *Resolver) DocumentationSearch(ctx context.Context, args *gql.DocumentationSearchArgs) (gql.DocumentationSearchResultsResolver, error) {
	var repos []string
	if args.Repos != nil {
		repos = *args.Repos
	}
	results, err := r.resolver.DocumentationSearch(ctx, args.Query, repos)
	if err != nil {
		return nil, err
	}
	return &documentationSearchResultsResolver{results: results}, nil
}

type documentationSearchResultsResolver struct {
	results []precise.DocumentationSearchResult
}

func (r *documentationSearchResultsResolver) Results() []gql.DocumentationSearchResultResolver {
	resolvers := make([]gql.DocumentationSearchResultResolver, 0, len(r.results))
	for _, result := range r.results {
		resolvers = append(resolvers, &documentationResultResolver{result: result})
	}
	return resolvers
}

type documentationResultResolver struct {
	result precise.DocumentationSearchResult
}

func (r *documentationResultResolver) Lang() string      { return r.result.Lang }
func (r *documentationResultResolver) RepoName() string  { return r.result.RepoName }
func (r *documentationResultResolver) SearchKey() string { return r.result.SearchKey }
func (r *documentationResultResolver) PathID() string    { return r.result.PathID }
func (r *documentationResultResolver) Label() string     { return r.result.Label }
func (r *documentationResultResolver) Detail() string    { return r.result.Detail }
func (r *documentationResultResolver) Tags() []string    { return r.result.Tags }
