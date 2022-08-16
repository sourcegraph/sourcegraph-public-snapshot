package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.SearchQueryInsightsResult = &searchQueryInsightUnionResolver{}

type searchQueryInsightsResolver struct{}

func (r *searchQueryInsightsResolver) Preview(ctx context.Context) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, nil
}

func (r *searchQueryInsightsResolver) ThirtyDayPercentChange(ctx context.Context) (int32, error) {
	return 0, nil
}

func newSearchQueryInsightUnionResolver() (graphqlbackend.SearchQueryInsightsResult, error) {
	// TODO(chwarwick): Replace with logic to determine if insights available for query
	return &searchQueryInsightUnionResolver{
		resolver: &insightsNotAvailable{},
	}, nil
}

// A  type to represent the GraphQL union SearchQueryInsightResult
type searchQueryInsightUnionResolver struct {
	resolver any
}

// ToQueryInsightPreviewSeries is used by the GraphQL library to resolve type fragments for unions
func (r *searchQueryInsightUnionResolver) ToSearchQueryInsights() (graphqlbackend.SearchQueryInsightsResolver, bool) {
	res, ok := r.resolver.(*searchQueryInsightsResolver)
	return res, ok
}

// ToQueryInsightNotAvailable is used by the GraphQL library to resolve type fragments for unions
func (r *searchQueryInsightUnionResolver) ToSearchQueryInsightsNotAvailable() (graphqlbackend.SearchQueryInsightsNotAvailable, bool) {
	res, ok := r.resolver.(*insightsNotAvailable)
	return res, ok
}

type insightsNotAvailable struct {
}

func (r *insightsNotAvailable) Message(ctx context.Context) string {
	return "no insights available for this query"
}
