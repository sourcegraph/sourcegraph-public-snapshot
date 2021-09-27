package graphqlbackend

import (
	"context"
	"math"

	"github.com/cockroachdb/errors"
)

var _ = SearchSuggestionResolver(&apiDocsSearchSuggestionResolver{})

type apiDocsSearchSuggestionResolver struct {
	baseSuggestionResolver
	score int
	r     DocumentationSearchResultResolver
}

// Implements GraphQL APIDocsSearchSuggestion resolver.
func (r *apiDocsSearchSuggestionResolver) Lang() string       { return r.r.Lang() }
func (r *apiDocsSearchSuggestionResolver) RepoName() string   { return r.r.RepoName() }
func (r *apiDocsSearchSuggestionResolver) SearchKey() string  { return r.r.SearchKey() }
func (r *apiDocsSearchSuggestionResolver) PathID() string     { return r.r.PathID() }
func (r *apiDocsSearchSuggestionResolver) NodeLabel() string  { return r.r.Label() }
func (r *apiDocsSearchSuggestionResolver) NodeDetail() string { return r.r.Detail() }
func (r *apiDocsSearchSuggestionResolver) Tags() []string     { return r.r.Tags() }

// Implements SearchSuggestionResolver interface
func (r *apiDocsSearchSuggestionResolver) Score() int    { return r.score }
func (r *apiDocsSearchSuggestionResolver) Length() int   { return len(r.r.SearchKey()) }
func (r *apiDocsSearchSuggestionResolver) Label() string { return r.r.SearchKey() }
func (r *apiDocsSearchSuggestionResolver) ToAPIDocsSearchSuggestion() (*apiDocsSearchSuggestionResolver, bool) {
	return r, true
}

// Implements SearchSuggestionResolver interface
func (r *apiDocsSearchSuggestionResolver) Key() suggestionKey {
	return suggestionKey{
		lang:          r.r.Lang(),
		repoName:      r.r.RepoName(),
		apiDocsPathID: r.r.PathID(),
	}
}

var mockShowAPIDocsSuggestions showSearchSuggestionResolvers

func (r *searchResolver) showAPIDocsSuggestions(ctx context.Context) ([]SearchSuggestionResolver, error) {
	if mockShowAPIDocsSuggestions != nil {
		return mockShowAPIDocsSuggestions()
	}

	search, err := r.codeIntelResolver.DocumentationSearch(ctx, &DocumentationSearchArgs{
		Query: r.rawQuery(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "DocumentationSearch")
	}
	results := search.Results()

	resolvers := make([]SearchSuggestionResolver, 0, len(results))
	for i, result := range results {
		resolvers = append(resolvers, &apiDocsSearchSuggestionResolver{
			score: math.MaxInt32-i,
			r:     result,
		})
	}
	return resolvers, nil
}
