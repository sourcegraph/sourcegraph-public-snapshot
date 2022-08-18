package resolvers

import (
	"context"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/search/query"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type searchAggregateResolver struct {
	baseInsightResolver
	searchQuery string
	patternType string
}

func (r *searchAggregateResolver) ModeAvailability(ctx context.Context) []graphqlbackend.AggregationModeAvailabilityResolver {
	resolvers := []graphqlbackend.AggregationModeAvailabilityResolver{}
	for _, mode := range types.SearchAggregationModes {
		resolvers = append(resolvers, newAggregationModeAvailabilityResolver(r.searchQuery, r.patternType, mode))
	}
	return resolvers
}

func (r *searchAggregateResolver) Aggregations(ctx context.Context, args graphqlbackend.AggregationsArgs) (graphqlbackend.SearchAggregationResultResolver, error) {
	// TODO(chwarwick): Replace with logic to detmine if available and return values
	return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver()}, nil
}

func newAggregationModeAvailabilityResolver(searchQuery string, patternType string, mode types.SearchAggregationMode) graphqlbackend.AggregationModeAvailabilityResolver {
	return &aggregationModeAvailabilityResolver{searchQuery: searchQuery, patternType: patternType, mode: mode}
}

type aggregationModeAvailabilityResolver struct {
	searchQuery string
	patternType string
	mode        types.SearchAggregationMode
}

func (r *aggregationModeAvailabilityResolver) Mode() string {
	return string(r.mode)
}

func (r *aggregationModeAvailabilityResolver) Available() (bool, error) {
	checkByMode := map[types.SearchAggregationMode]canAggregateBy{
		types.REPO_AGGREGATION_MODE: canAggregateByRepo,
		// types.PATH_AGGREGATION_MODE: canAggregateByPath,
	}
	canAggregateByFunc, ok := checkByMode[r.mode]
	if !ok {
		return false, errors.Newf("mode %q not recognised", r.mode)
	}
	return canAggregateByFunc(r.searchQuery, r.patternType), nil
}

type canAggregateBy func(searchQuery, patternType string) bool

func canAggregateByRepo(searchQuery, patternType string) bool {
	// We can always aggregate by repo.
	return true
}

func canAggregateByPath(searchQuery, patternType string) bool {
	// We don't need to validate the searchQuery is valid as search have their own validation which would get triggered,
	// we just want to grab the query plan parameters.
	plan, _ := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	parameters := querybuilder.ParametersFromQueryPlan(plan)
	// cannot aggregate over:
	// - searches by commit or repo
	for _, parameter := range parameters {
		if parameter.Field == query.FieldSelect || parameter.Field == query.FieldType {
			if parameter.Value == "commit" || parameter.Value == "repo" {
				return false
			}
		}
	}
	return true
}

func (r *aggregationModeAvailabilityResolver) ReasonUnavailable() (*string, error) {
	reason := "not implemented"
	return &reason, nil
}

// A  type to represent the GraphQL union SearchAggregationResult
type searchAggregationResultResolver struct {
	resolver any
}

// ToExhaustiveSearchAggregationResult is used by the GraphQL library to resolve type fragments for unions
func (r *searchAggregationResultResolver) ToExhaustiveSearchAggregationResult() (graphqlbackend.ExhaustiveSearchAggregationResultResolver, bool) {
	res, ok := r.resolver.(*searchAggregationModeResultResolver)
	if ok && res.isExhaustive {
		return res, ok
	}
	return nil, false
}

// ToNonExhaustiveSearchAggregationResult is used by the GraphQL library to resolve type fragments for unions
func (r *searchAggregationResultResolver) ToNonExhaustiveSearchAggregationResult() (graphqlbackend.NonExhaustiveSearchAggregationResultResolver, bool) {
	res, ok := r.resolver.(*searchAggregationModeResultResolver)
	if ok && !res.isExhaustive {
		return res, ok
	}
	return nil, false
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

// Resolver to calculate aggregations for a combination of search query, pattern type, aggregation mode
type searchAggregationModeResultResolver struct {
	baseInsightResolver
	searchQuery  string
	patternType  string
	mode         types.SearchAggregationMode
	isExhaustive bool
}

func (r *searchAggregationModeResultResolver) Groups() ([]graphqlbackend.AggregationGroup, error) {
	return nil, errors.New("not implemented")
}

func (r *searchAggregationModeResultResolver) OtherResultCount() (*int32, error) {
	return nil, errors.New("not implemented")
}

// OtherGroupCount - used for exhaustive aggregations to indicate count of additional groups
func (r *searchAggregationModeResultResolver) OtherGroupCount() (*int32, error) {
	return nil, errors.New("not implemented")
}

// ApproximateOtherGroupCount - used for nonexhaustive aggregations to indicate approx count of additional groups
func (r *searchAggregationModeResultResolver) ApproximateOtherGroupCount() (*int32, error) {
	return nil, errors.New("not implemented")
}

func (r *searchAggregationModeResultResolver) SupportsPersistence() (*bool, error) {
	supported := false
	return &supported, nil
}
