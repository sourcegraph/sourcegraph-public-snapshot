package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
)

var _ graphqlbackend.QueryInsightsResult = &queryInsightUnionResolver{}

type queryInsightsResolver struct {
	query       string
	patternType searchquery.SearchType

	baseInsightResolver
}

func (r *queryInsightsResolver) Preview(ctx context.Context) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, nil
}

func (r *queryInsightsResolver) ThirtyDayPercentChange(ctx context.Context) (int32, error) {
	return 0, nil
}

func newQueryInsightResolver(searchQuery, patternType string) (queryInsightsResolver, error) {
	var searchType query.SearchType
	switch patternType {
	case "literal":
		searchType = searchquery.SearchTypeLiteral
	case "structural":
		searchType = searchquery.SearchTypeStructural
	case "regexp", "regex":
		searchType = searchquery.SearchTypeRegex
	default:
		searchType = searchquery.SearchTypeLiteral
	}
	return queryInsightsResolver{query: searchQuery, patternType: searchType}, nil

}

func newQueryInsightUnionResolver(searchQuery, patternType string) (graphqlbackend.QueryInsightsResult, error) {
	// TODO(chwarwick): Replace with logic to determine if insights available for query
	return &queryInsightUnionResolver{
		resolver: &insightsNotAvailable{},
	}, nil
}

// A  type to represent the GraphQL union QueryInsightResult
type queryInsightUnionResolver struct {
	resolver any
}

// ToQueryInsightPreviewSeries is used by the GraphQL library to resolve type fragments for unions
func (r *queryInsightUnionResolver) ToQueryInsights() (graphqlbackend.QueryInsightsResolver, bool) {
	res, ok := r.resolver.(*queryInsightsResolver)
	return res, ok
}

// ToQueryInsightNotAvailable is used by the GraphQL library to resolve type fragments for unions
func (r *queryInsightUnionResolver) ToQueryInsightsNotAvailable() (graphqlbackend.QueryInsightsNotAvailable, bool) {
	res, ok := r.resolver.(*insightsNotAvailable)
	return res, ok
}

type insightsNotAvailable struct {
}

func (r *insightsNotAvailable) Message(ctx context.Context) string {
	return "no insights available for this query"
}
