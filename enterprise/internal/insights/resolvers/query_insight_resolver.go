package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
)

var _ graphqlbackend.QueryInsightResolver = &queryInsightResolver{}

type queryInsightResolver struct {
	query       string
	patternType searchquery.SearchType

	baseInsightResolver
}

func (r *queryInsightResolver) InsightPreview(ctx context.Context) (graphqlbackend.QueryInsightPreview, error) {
	return nil, nil
}

func (r *queryInsightResolver) ThirtyDayChange(ctx context.Context) (graphqlbackend.QueryInsightResultChange, error) {
	return nil, nil
}

func newQueryInsightResolver(searchQuery, patternType string) (queryInsightResolver, error) {
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
	return queryInsightResolver{query: searchQuery, patternType: searchType}, nil

}

// A dummy type to represent the GraphQL union QueryInsightPreview
type insightThirtyDayChangeUnionResolver struct {
	resolver any
}

// ToQueryInsightPreviewSeries is used by the GraphQL library to resolve type fragments for unions
func (r *insightThirtyDayChangeUnionResolver) ToResultChange() (graphqlbackend.ResultChange, bool) {
	res, ok := r.resolver.(*thirtyDayChange)
	return res, ok
}

// ToQueryInsightNotAvailable is used by the GraphQL library to resolve type fragments for unions
func (r *insightThirtyDayChangeUnionResolver) ToQueryInsightNotAvailable() (graphqlbackend.QueryInsightNotAvailable, bool) {
	res, ok := r.resolver.(*insightNotAvailable)
	return res, ok
}

// A dummy type to represent the GraphQL union QueryInsightPreview
type insightQueryPreviewUnionResolver struct {
	resolver any
}

// ToQueryInsightPreviewSeries is used by the GraphQL library to resolve type fragments for unions
func (r *insightPresentationUnionResolver) ToQueryInsightPreviewSeries() (graphqlbackend.QueryInsightPreviewSeries, bool) {
	res, ok := r.resolver.(*queryInsightPreview)
	return res, ok
}

// ToQueryInsightNotAvailable is used by the GraphQL library to resolve type fragments for unions
func (r *insightPresentationUnionResolver) ToQueryInsightNotAvailable() (graphqlbackend.QueryInsightNotAvailable, bool) {
	res, ok := r.resolver.(*insightNotAvailable)
	return res, ok
}

type insightNotAvailable struct {
}

func (r *insightNotAvailable) Message(ctx context.Context) string {
	return "insight not available for query"
}

type thirtyDayChange struct {
}

func (r *thirtyDayChange) Current(ctx context.Context) int32 {
	return 0
}
func (r *thirtyDayChange) Previous(ctx context.Context) int32 {
	return 0
}
func (r *thirtyDayChange) PercentChange(ctx context.Context) int32 {
	return 0
}

type queryInsightPreview struct {
}

func (r *queryInsightPreview) Series(ctx context.Context) ([]graphqlbackend.SearchInsightLivePreviewSeriesResolver, error) {
	return nil, nil
}
