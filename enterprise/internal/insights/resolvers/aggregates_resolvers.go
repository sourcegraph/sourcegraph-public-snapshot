package resolvers

import (
	"context"
	"fmt"
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
	searchQuery       string
	patternType       string
	mode              types.SearchAggregationMode
	reasonUnavailable *string
}

func (r *aggregationModeAvailabilityResolver) Mode() string {
	return string(r.mode)
}

func (r *aggregationModeAvailabilityResolver) Available() (bool, error) {
	checkByMode := map[types.SearchAggregationMode]canAggregateBy{
		types.REPO_AGGREGATION_MODE: canAggregateByRepo,
		// TODO(insights): these paths should be uncommented as they are implemented. Logic for allowing the aggregation should be double-checked.
		// types.PATH_AGGREGATION_MODE: canAggregateByPath,
		// types.AUTHOR_AGGREGATION_MODE: canAggregateByAuthor,
		// types.CAPTURE_GROUP_AGGREGATION_MODE: canAggregateByCaptureGroup,
	}
	canAggregateByFunc, ok := checkByMode[r.mode]
	if !ok {
		reason := fmt.Sprintf("aggregation mode %v is not yet supported", r.mode)
		r.reasonUnavailable = &reason
		return false, nil
	}
	canAggregate, err := canAggregateByFunc(r.searchQuery, r.patternType)
	if err != nil {
		reason := fmt.Sprintf("cannot aggregate due to error: %v", err)
		r.reasonUnavailable = &reason
	}
	if !canAggregate {
		reason := fmt.Sprintf("this specific query does not support aggregation by %v", r.mode)
		r.reasonUnavailable = &reason
	}
	return canAggregate, err
}

func (r *aggregationModeAvailabilityResolver) ReasonUnavailable() *string {
	if r.reasonUnavailable != nil {
		return r.reasonUnavailable
	}
	return nil
}

type canAggregateBy func(searchQuery, patternType string) (bool, error)

func canAggregateByRepo(searchQuery, patternType string) (bool, error) {
	_, err := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseAndValidateQuery")
	}
	// We can always aggregate by repo.
	return true, nil
}

func canAggregateByPath(searchQuery, patternType string) (bool, error) {
	plan, err := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseAndValidateQuery")
	}
	parameters := querybuilder.ParametersFromQueryPlan(plan)
	// cannot aggregate over:
	// - searches by commit or repo
	for _, parameter := range parameters {
		if parameter.Field == query.FieldSelect || parameter.Field == query.FieldType {
			if parameter.Value == "commit" || parameter.Value == "repo" {
				return false, nil
			}
		}
	}
	return true, nil
}

func canAggregateByAuthor(searchQuery, patternType string) (bool, error) {
	plan, err := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseAndValidateQuery")
	}
	parameters := querybuilder.ParametersFromQueryPlan(plan)
	// can only aggregate over type:diff and select/type:commit searches.
	// users can make searches like `type:commit fix select:repo` but assume a faulty search like that is on them.
	for _, parameter := range parameters {
		if parameter.Field == query.FieldSelect || parameter.Field == query.FieldType {
			if parameter.Value == "diff" || parameter.Value == "commit" {
				return true, nil
			}
		}
	}
	return false, nil
}

func canAggregateByCaptureGroup(searchQuery, patternType string) (bool, error) {
	if !(patternType == "regexp" || patternType == "regex" || patternType == "standard") {
		return false, nil
	}
	plan, err := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseAndValidateQuery")
	}
	parameters := querybuilder.ParametersFromQueryPlan(plan)
	selectParameter, typeParameter := false, false
	for _, parameter := range parameters {
		if parameter.Field == query.FieldSelect {
			if parameter.Value == "repo" || parameter.Value == "file" {
				selectParameter = true
			}
		} else if parameter.Field == query.FieldType {
			if parameter.Value == "repo" || parameter.Value == "path" {
				typeParameter = true
			}
		}
	}
	if selectParameter && !typeParameter {
		return false, nil
	}
	return true, nil
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
