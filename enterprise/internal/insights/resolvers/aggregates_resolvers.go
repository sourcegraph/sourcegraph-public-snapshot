package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SearchAggregationMode string

const (
	REPO_AGGREGATION_MODE          SearchAggregationMode = "REPO"
	PATH_AGGREGATION_MODE          SearchAggregationMode = "PATH"
	AUTHOR_AGGREGATION_MODE        SearchAggregationMode = "AUTHOR"
	CAPTURE_GROUP_AGGREGATION_MODE SearchAggregationMode = "CAPTURE_GROUP"
)

var SearchAggregationModes = []SearchAggregationMode{REPO_AGGREGATION_MODE, PATH_AGGREGATION_MODE, AUTHOR_AGGREGATION_MODE, CAPTURE_GROUP_AGGREGATION_MODE}

type searchAggregateResolver struct {
	baseInsightResolver
	searchQuery string
	patternType query.SearchType
}

func (r *searchAggregateResolver) ModeAvailability(ctx context.Context) []graphqlbackend.AggregationModeAvailabilityResolver {
	resolvers := []graphqlbackend.AggregationModeAvailabilityResolver{}
	for _, mode := range SearchAggregationModes {
		resolvers = append(resolvers, newAggregationModeAvailabilityResolver(r.searchQuery, r.patternType, mode))
	}
	return resolvers
}

func (r *searchAggregateResolver) Aggregations(ctx context.Context, args graphqlbackend.AggregationsArgs) (graphqlbackend.SearchAggregationResultResolver, error) {
	// TODO(chwarwick): Replace with logic to detmine if available and return values
	return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver()}, nil
}

func newAggregationModeAvailabilityResolver(searchQuery string, patternType query.SearchType, mode SearchAggregationMode) graphqlbackend.AggregationModeAvailabilityResolver {
	return &aggregationModeAvailabilityResolver{searchQuery: searchQuery, patternType: patternType, mode: mode}
}

type aggregationModeAvailabilityResolver struct {
	searchQuery string
	patternType query.SearchType
	mode        SearchAggregationMode
}

func (r *aggregationModeAvailabilityResolver) Mode() string {
	return string(r.mode)
}

func (r *aggregationModeAvailabilityResolver) Available() (bool, error) {
	return false, nil
}

func (r *aggregationModeAvailabilityResolver) ReasonUnavailable() (*string, error) {
	reason := "not implemented"
	return &reason, nil
}

// A  type to represent the GraphQL union SearchAggregationResult
type searchAggregationResultResolver struct {
	resolver any
}

// ToSearchAggregationModeResult is used by the GraphQL library to resolve type fragments for unions
func (r *searchAggregationResultResolver) ToSearchAggregationModeResult() (graphqlbackend.SearchAggregationModeResultResolver, bool) {
	res, ok := r.resolver.(*searchAggregationModeResultResolver)
	return res, ok
}

// ToSearchAggregationNotAvailable is used by the GraphQL library to resolve type fragments for unions
func (r *searchAggregationResultResolver) ToSearchAggregationNotAvailable() (graphqlbackend.SearchAggregationNotAvailable, bool) {
	res, ok := r.resolver.(*searchAggregationNotAvailableResolver)
	return res, ok
}

func newSearchAggregationNotAvailableResolver() graphqlbackend.SearchAggregationNotAvailable {
	return &searchAggregationNotAvailableResolver{}
}

type searchAggregationNotAvailableResolver struct {
}

func (r *searchAggregationNotAvailableResolver) Reason() string {
	return "not implemented"
}

// Resolver to calcuate aggregations for a combination of search query, pattern type, aggregation mode
type searchAggregationModeResultResolver struct {
	baseInsightResolver
	searchQuery string
	patternType query.SearchType
	mode        SearchAggregationMode
}

func (r *searchAggregationModeResultResolver) Values() ([]graphqlbackend.AggregationValue, error) {
	return nil, errors.New("not implemented")
}

func (r *searchAggregationModeResultResolver) OtherResultCount() (*int32, error) {
	return nil, errors.New("not implemented")
}

func (r *searchAggregationModeResultResolver) OtherValueCount() (*int32, error) {
	return nil, errors.New("not implemented")
}

func (r *searchAggregationModeResultResolver) IsExhaustive() (*bool, error) {
	return nil, errors.New("not implemented")
}
