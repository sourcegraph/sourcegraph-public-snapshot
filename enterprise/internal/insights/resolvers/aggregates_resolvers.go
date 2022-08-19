package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const searchTimeLimitSeconds = 2

//TODO: move to a setting
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
	supported, reason := aggregationModeSupported(r.searchQuery, r.patternType, aggregationMode)
	if !supported {
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(reason, aggregationMode)}, nil
	}

	// If a search includes a timeout it reports as completing succesfully with the timeout is hit
	// This includes a timeout in the search that is a second longer than the context we will cancel as a fail safe
	modifiedQuery, err := querybuilder.AggregationQuery(querybuilder.BasicQuery(r.searchQuery), searchTimeLimitSeconds+1)
	if err != nil {
		return &searchAggregationResultResolver{
			resolver: newSearchAggregationNotAvailableResolver("search query could not be expanded for aggregation", aggregationMode),
		}, nil
	}

	cappedAggregator := streaming.NewLimitedAggregator(aggregationBufferSize)
	tabulationErrors := []error{}
	tabulationFunc := func(amr *streaming.AggregationMatchResult, err error) {
		if err != nil {
			tabulationErrors = append(tabulationErrors, err)
			return
		}
		cappedAggregator.Add(amr.Key.Group, int32(amr.Count))
	}
	onMatchFunc, err := streaming.TabulateAggregationMatches(tabulationFunc, aggregationMode)
	if err != nil {
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(err.Error(), aggregationMode)}, nil
	}
	decoder, searchEvents := streaming.AggregationDecoder(onMatchFunc)

	// This request context limits the total time search runs and returns in a manner that allows knowing if the search completed
	requestContext, cancelReqContext := context.WithTimeout(ctx, time.Second*searchTimeLimitSeconds)
	defer cancelReqContext()
	err = streaming.Search(requestContext, string(modifiedQuery), &r.patternType, decoder)
	if err != nil {
		reason := "unable to run search"
		if errors.Is(err, context.DeadlineExceeded) {
			reason = "search did not complete in time"
		}
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(reason, aggregationMode)}, nil
	}

	successful, failureReason := searchSuccessful(searchEvents, tabulationErrors)
	if !successful {
		return &searchAggregationResultResolver{resolver: newSearchAggregationNotAvailableResolver(failureReason, aggregationMode)}, nil
	}

	results := buildResults(cappedAggregator, int(args.Limit), aggregationMode, r.searchQuery)

	return &searchAggregationResultResolver{resolver: &searchAggregationModeResultResolver{
		baseInsightResolver: r.baseInsightResolver,
		searchQuery:         r.searchQuery,
		patternType:         r.patternType,
		mode:                aggregationMode,
		results:             results,
		isExhaustive:        cappedAggregator.OtherCounts().GroupCount == 0,
	}}, nil

}

// Temporary placeholder to limit mode to only REPO
func aggregationModeSupported(searchQuery, patternType string, mode types.SearchAggregationMode) (bool, string) {
	if mode == types.REPO_AGGREGATION_MODE {
		return true, ""
	}
	return false, "Only aggregation by repository is currently supported."
}

func searchSuccessful(events *streaming.AggregationDecoderEvents, tabulationErrors []error) (bool, string) {
	for _, skipped := range events.Skipped {
		if skipped.Reason == string(streaming.ShardTimeoutSkippedReason) {
			return false, "query was unable to complete"
		}
	}

	if len(events.Errors) > 0 {
		return false, "query returned with errors"
	}

	if len(tabulationErrors) > 0 {
		return false, "query returned with errors"
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

func buildResults(aggregator streaming.LimitedAggregator, limit int, mode types.SearchAggregationMode, originalQuery string) aggregationResults {
	sorted := aggregator.SortAggregate()
	groups := make([]graphqlbackend.AggregationGroup, 0, limit)
	otherResults := aggregator.OtherCounts().ResultCount
	otherGroups := aggregator.OtherCounts().GroupCount

	for i := 0; i < len(sorted); i++ {
		if i < limit {
			groups = append(groups, &AggregationGroup{
				label: sorted[i].Label,
				count: int(sorted[i].Count),
				query: nil,
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
