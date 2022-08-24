package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/search/client"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/aggregation"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const searchTimeLimitSeconds = 2

// TODO: move to a setting
const aggregationBufferSize = 500

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
	// Steps:
	// 1. - If no mode get the default mode (currently defaulted in gql to REPO)
	// 2. - Validate mode can be used supported
	// 3. - Modify search query (timeout: & count:)
	// 3. - Run Search
	// 4. - Check search for errors/alerts
	// 5 -  Generate correct resolver pass search results if valid
	aggregationMode := types.SearchAggregationMode(args.Mode)
	aggregationModeAvailabilityResolver := newAggregationModeAvailabilityResolver(r.searchQuery, r.patternType, aggregationMode)
	supported, err := aggregationModeAvailabilityResolver.Available()
	if err != nil {
		return nil, err
	}
	if !supported {
		unavailableReason := ""
		// We don't need to assert on the error because this uses the same logic as `Available()` above so it would
		// have errored already.
		reason, _ := aggregationModeAvailabilityResolver.ReasonUnavailable()
		if reason == nil {
			unavailableReason = "could not fetch unavailability reason"
		} else {
			unavailableReason = *reason
		}
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(unavailableReason, aggregationMode)}, nil
	}

	// If a search includes a timeout it reports as completing succesfully with the timeout is hit
	// This includes a timeout in the search that is a second longer than the context we will cancel as a fail safe
	modifiedQuery, err := querybuilder.AggregationQuery(querybuilder.BasicQuery(r.searchQuery), searchTimeLimitSeconds+1)
	if err != nil {
		return &searchAggregationResultResolver{
			resolver: newSearchAggregationNotAvailableResolver("search query could not be expanded for aggregation", aggregationMode),
		}, nil
	}

	cappedAggregator := aggregation.NewLimitedAggregator(aggregationBufferSize)
	tabulationErrors := []error{}
	tabulationFunc := func(amr *aggregation.AggregationMatchResult, err error) {
		if err != nil {
			tabulationErrors = append(tabulationErrors, err)
			return
		}
		cappedAggregator.Add(amr.Key.Group, int32(amr.Count))
	}

	countingFunc, err := aggregation.GetCountFuncForMode(r.searchQuery, r.patternType, aggregationMode)
	if err != nil {
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(err.Error(), aggregationMode)}, nil
	}

	requestContext, cancelReqContext := context.WithTimeout(ctx, time.Second*searchTimeLimitSeconds)
	defer cancelReqContext()
	searchClient := streaming.NewInsightsSearchClient(r.baseInsightResolver.postgresDB)
	searchResultsAggregator := aggregation.NewSearchResultsAggregator(tabulationFunc, countingFunc)
	alert, err := searchClient.Search(requestContext, string(modifiedQuery), &r.patternType, searchResultsAggregator)
	if err != nil || requestContext.Err() != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestContext.Err(), context.DeadlineExceeded) {
			return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver("query unable to complete in allocated time", aggregationMode)}, nil
		} else {
			return nil, err
		}
	}

	successful, failureReason := searchSuccessful(alert, tabulationErrors, searchResultsAggregator.ShardTimeoutOccurred())
	if !successful {
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(failureReason, aggregationMode)}, nil
	}

	results := buildResults(cappedAggregator, int(args.Limit), aggregationMode, r.searchQuery, r.patternType)

	return &searchAggregationResultResolver{resolver: &searchAggregationModeResultResolver{
		baseInsightResolver: r.baseInsightResolver,
		searchQuery:         r.searchQuery,
		patternType:         r.patternType,
		mode:                aggregationMode,
		results:             results,
		isExhaustive:        cappedAggregator.OtherCounts().GroupCount == 0,
	}}, nil
}

func searchSuccessful(alert *search.Alert, tabulationErrors []error, shardTimeoutOccurred bool) (bool, string) {

	if len(tabulationErrors) > 0 {
		return false, "query returned with errors"
	}

	if shardTimeoutOccurred {
		return false, "query unable to complete in allocated time"
	}

	return true, ""
}

type aggregationResults struct {
	groups           []graphqlbackend.AggregationGroup
	otherResultCount int
	otherGroupCount  int
}

type AggregationGroup struct {
	label string
	count int
	query *string
}

func (r *AggregationGroup) Label() string {
	return r.label
}
func (r *AggregationGroup) Count() int32 {
	return int32(r.count)
}
func (r *AggregationGroup) Query() (*string, error) {
	return r.query, nil
}

func buildResults(aggregator aggregation.LimitedAggregator, limit int, mode types.SearchAggregationMode, originalQuery string, patternType string) aggregationResults {
	sorted := aggregator.SortAggregate()
	groups := make([]graphqlbackend.AggregationGroup, 0, limit)
	otherResults := aggregator.OtherCounts().ResultCount
	otherGroups := aggregator.OtherCounts().GroupCount

	for i := 0; i < len(sorted); i++ {
		if i < limit {
			label := sorted[i].Label
			drilldownQuery, err := buildDrilldownQuery(mode, originalQuery, label, patternType)
			if err != nil {
				// for some reason we couldn't generate a new query, so fallback to the original
				drilldownQuery = originalQuery
			}
			groups = append(groups, &AggregationGroup{
				label: label,
				count: int(sorted[i].Count),
				query: &drilldownQuery,
			})
		} else {
			otherGroups++
			otherResults += sorted[i].Count
		}
	}

	return aggregationResults{
		groups:           groups,
		otherResultCount: int(otherResults),
		otherGroupCount:  int(otherGroups),
	}
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
	canAggregateByFunc := getAggregateBy(r.mode)
	if canAggregateByFunc == nil {
		return false, nil
	}
	return canAggregateByFunc(r.searchQuery, r.patternType)
}

func (r *aggregationModeAvailabilityResolver) ReasonUnavailable() (*string, error) {
	// if it’s possible write a clear concise reason why that mode won’t work then put it in the reason.
	// if not return an error
	canAggregateByFunc := getAggregateBy(r.mode)
	if canAggregateByFunc == nil {
		reason := fmt.Sprintf("aggregation mode %v is not yet supported", r.mode)
		return &reason, nil
	}
	canAggregate, err := canAggregateByFunc(r.searchQuery, r.patternType)
	if err != nil {
		return nil, err
	}
	if !canAggregate {
		reason := fmt.Sprintf("this specific query does not support aggregation by %v", r.mode)
		return &reason, nil
	}
	return nil, nil
}

func getAggregateBy(mode types.SearchAggregationMode) canAggregateBy {
	checkByMode := map[types.SearchAggregationMode]canAggregateBy{
		types.REPO_AGGREGATION_MODE:   canAggregateByRepo,
		types.PATH_AGGREGATION_MODE:   canAggregateByPath,
		types.AUTHOR_AGGREGATION_MODE: canAggregateByAuthor,
		// TODO(insights): these paths should be uncommented as they are implemented. Logic for allowing the aggregation should be double-checked.
		// types.CAPTURE_GROUP_AGGREGATION_MODE: canAggregateByCaptureGroup,
	}
	canAggregateByFunc, ok := checkByMode[mode]
	if !ok {
		return nil
	}
	return canAggregateByFunc
}

type canAggregateBy func(searchQuery, patternType string) (bool, error)

func canAggregateByRepo(searchQuery, patternType string) (bool, error) {
	_, err := querybuilder.ParseQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseQuery")
	}
	// We can always aggregate by repo.
	return true, nil
}

func canAggregateByPath(searchQuery, patternType string) (bool, error) {
	plan, err := querybuilder.ParseQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseQuery")
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
	plan, err := querybuilder.ParseQuery(searchQuery, patternType)
	if err != nil {
		return false, errors.Wrapf(err, "ParseQuery")
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
	// TODO(leonore): Finish up logic for ability to aggregate by capture group.
	// A query should contain a capture group to allow this kind of aggregation.
	// if !(patternType == "regexp" || patternType == "regex" || patternType == "standard" || patternType == "lucky") {
	//	return false, nil
	// }
	// plan, err := querybuilder.ParseAndValidateQuery(searchQuery, patternType)
	// if err != nil {
	//	return false, errors.Wrapf(err, "ParseAndValidateQuery")
	// }
	// parameters := querybuilder.ParametersFromQueryPlan(plan)
	// selectParameter, typeParameter := false, false
	// for _, parameter := range parameters {
	//	if parameter.Field == query.FieldSelect {
	//		if parameter.Value == "repo" || parameter.Value == "file" {
	//			selectParameter = true
	//		}
	//	} else if parameter.Field == query.FieldType {
	//		if parameter.Value == "repo" || parameter.Value == "path" {
	//			typeParameter = true
	//		}
	//	}
	// }
	// if selectParameter && !typeParameter {
	//	return false, nil
	// }
	// return true, nil
	return false, nil
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

func newSearchAggregationNotAvailableResolver(reason string, mode types.SearchAggregationMode) graphqlbackend.SearchAggregationNotAvailable {
	return &searchAggregationNotAvailableResolver{
		reason: reason,
		mode:   mode,
	}
}

type searchAggregationNotAvailableResolver struct {
	reason string
	mode   types.SearchAggregationMode
}

func (r *searchAggregationNotAvailableResolver) Reason() string {
	return r.reason
}
func (r *searchAggregationNotAvailableResolver) Mode() string {
	return string(r.mode)
}

// Resolver to calculate aggregations for a combination of search query, pattern type, aggregation mode
type searchAggregationModeResultResolver struct {
	baseInsightResolver
	searchQuery  string
	patternType  string
	mode         types.SearchAggregationMode
	results      aggregationResults
	isExhaustive bool
}

func (r *searchAggregationModeResultResolver) Groups() ([]graphqlbackend.AggregationGroup, error) {
	return r.results.groups, nil
}

func (r *searchAggregationModeResultResolver) OtherResultCount() (*int32, error) {
	var count int32 = int32(r.results.otherResultCount)
	return &count, nil
}

// OtherGroupCount - used for exhaustive aggregations to indicate count of additional groups
func (r *searchAggregationModeResultResolver) OtherGroupCount() (*int32, error) {
	var count int32 = int32(r.results.otherGroupCount)
	return &count, nil
}

// ApproximateOtherGroupCount - used for nonexhaustive aggregations to indicate approx count of additional groups
func (r *searchAggregationModeResultResolver) ApproximateOtherGroupCount() (*int32, error) {
	var count int32 = int32(r.results.otherGroupCount)
	return &count, nil
}

func (r *searchAggregationModeResultResolver) SupportsPersistence() (*bool, error) {
	supported := false
	return &supported, nil
}

func (r *searchAggregationModeResultResolver) Mode() (string, error) {
	return string(r.mode), nil
}

func buildDrilldownQuery(mode types.SearchAggregationMode, originalQuery string, drilldown string, patternType string) (string, error) {
	var modifierFunc func(querybuilder.BasicQuery, string) (querybuilder.BasicQuery, error)
	switch mode {
	case types.REPO_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddRepoFilter
	case types.PATH_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddFileFilter
	case types.AUTHOR_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddAuthorFilter
	case types.CAPTURE_GROUP_AGGREGATION_MODE:
		searchType, err := client.SearchTypeFromString(patternType)
		if err != nil {
			return "", err
		}
		replacer, err := querybuilder.NewPatternReplacer(querybuilder.BasicQuery(originalQuery), searchType)
		if err != nil {
			return "", err
		}
		modifierFunc = func(basicQuery querybuilder.BasicQuery, s string) (querybuilder.BasicQuery, error) {
			return replacer.Replace(s)
		}
	default:
		return "", errors.New("unsupported aggregation mode")
	}

	newQuery, err := modifierFunc(querybuilder.BasicQuery(originalQuery), drilldown)
	return string(newQuery), err
}
