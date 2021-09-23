package graphqlbackend

import (
	"context"
	"math"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

var _ = SearchSuggestionResolver(&apiDocsSearchSuggestionResolver{})

type apiDocsSearchSuggestionResolver struct {
	baseSuggestionResolver
	score int
	r     precise.DocumentationSearchResult
}

// Implements GraphQL APIDocsSearchSuggestion resolver.
func (r *apiDocsSearchSuggestionResolver) Lang() string       { return r.r.Lang }
func (r *apiDocsSearchSuggestionResolver) RepoName() string   { return r.r.RepoName }
func (r *apiDocsSearchSuggestionResolver) SearchKey() string  { return r.r.SearchKey }
func (r *apiDocsSearchSuggestionResolver) PathID() string     { return r.r.PathID }
func (r *apiDocsSearchSuggestionResolver) NodeLabel() string  { return r.r.Label }
func (r *apiDocsSearchSuggestionResolver) NodeDetail() string { return r.r.Detail }
func (r *apiDocsSearchSuggestionResolver) Tags() []string     { return r.r.Tags }

// Implements SearchSuggestionResolver interface
func (r *apiDocsSearchSuggestionResolver) Score() int    { return r.score }
func (r *apiDocsSearchSuggestionResolver) Length() int   { return len(r.r.SearchKey) }
func (r *apiDocsSearchSuggestionResolver) Label() string { return r.r.SearchKey }
func (r *apiDocsSearchSuggestionResolver) ToAPIDocsSearchSuggestion() (*apiDocsSearchSuggestionResolver, bool) {
	return r, true
}

// Implements SearchSuggestionResolver interface
func (r *apiDocsSearchSuggestionResolver) Key() suggestionKey {
	return suggestionKey{
		lang:          r.r.Lang,
		repoName:      r.r.RepoName,
		apiDocsPathID: r.r.PathID,
	}
}

var mockShowAPIDocsSuggestions showSearchSuggestionResolvers

func (r *searchResolver) showAPIDocsSuggestions(ctx context.Context) ([]SearchSuggestionResolver, error) {
	if mockShowAPIDocsSuggestions != nil {
		return mockShowAPIDocsSuggestions()
	}

	var resolvers []SearchSuggestionResolver
	resolvers = append(resolvers, &apiDocsSearchSuggestionResolver{
		score: math.MaxInt32,
		r: precise.DocumentationSearchResult{
			Lang:      "go",
			RepoName:  "github.com/gorilla/mux",
			SearchKey: "mux.Router",
			PathID:    "/#Router",
			Label:     "type Router struct",
			Detail:    "Router keeps track of asdowiajdowiajdowijadoiwjad woaid jwoadij wdoiwjad oiwjaodiwj doiwjaodijwaod",
			Tags:      []string{"struct"},
		},
	})
	return resolvers, nil
}
