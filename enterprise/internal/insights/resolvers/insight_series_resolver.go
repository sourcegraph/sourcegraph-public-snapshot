package resolvers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	sctypes "github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
)

var _ graphqlbackend.InsightSeriesResolver = &precalculatedInsightSeriesResolver{}

// TODO(insights): remove insightSeriesResolver when `insights` is removed from graphql query schema
type insightSeriesResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	series          types.InsightViewSeries
	metadataStore   store.InsightMetadataStore

	filters types.InsightViewFilters
}

func (r *insightSeriesResolver) SeriesId() string { return r.series.SeriesID }

func (r *insightSeriesResolver) Label() string { return r.series.Label }

func (r *insightSeriesResolver) Points(ctx context.Context, _ *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var opts store.SeriesPointsOpts

	// Query data points only for the series we are representing.
	seriesID := r.series.SeriesID
	opts.SeriesID = &seriesID

	// Default to last 12 frames of data
	frames := query.BuildFrames(12, timeseries.TimeInterval{
		Unit:  types.IntervalUnit(r.series.SampleIntervalUnit),
		Value: r.series.SampleIntervalValue,
	}, time.Now())
	oldest := time.Now().AddDate(-1, 0, 0)
	if len(frames) != 0 {
		possibleOldest := frames[0].From
		if possibleOldest.Before(oldest) {
			oldest = possibleOldest
		}
	}
	opts.From = &oldest

	includeRepo := func(regex ...string) {
		opts.IncludeRepoRegex = append(opts.IncludeRepoRegex, regex...)
	}
	excludeRepo := func(regex ...string) {
		opts.ExcludeRepoRegex = append(opts.ExcludeRepoRegex, regex...)
	}

	if r.filters.IncludeRepoRegex != nil {
		includeRepo(*r.filters.IncludeRepoRegex)
	}
	if r.filters.ExcludeRepoRegex != nil {
		excludeRepo(*r.filters.ExcludeRepoRegex)
	}

	scLoader := &scLoader{primary: r.workerBaseStore.Handle()}
	inc, exc, err := unwrapSearchContexts(ctx, scLoader, r.filters.SearchContexts)
	if err != nil {
		return nil, errors.Wrap(err, "unwrapSearchContexts")
	}
	includeRepo(inc...)
	excludeRepo(exc...)

	points, err := r.insightsStore.SeriesPoints(ctx, opts)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightsDataPointResolver, 0, len(points))
	for _, point := range points {
		resolvers = append(resolvers, insightsDataPointResolver{point})
	}
	return resolvers, nil
}

// SearchContextLoader loads search contexts just from the full name of the
// context. This will not verify that the calling context owns the context, it
// will load regardless of the current user.
type SearchContextLoader interface {
	GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error)
}

type scLoader struct {
	primary dbutil.DB
}

func (l *scLoader) GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error) {
	db := database.NewUntypedDB(l.primary)
	return searchcontexts.ResolveSearchContextSpec(ctx, db, name)
}

func unwrapSearchContexts(ctx context.Context, loader SearchContextLoader, rawContexts []string) ([]string, []string, error) {
	var include []string
	var exclude []string

	for _, rawContext := range rawContexts {
		searchContext, err := loader.GetByName(ctx, rawContext)
		if err != nil {
			return nil, nil, err
		}
		if searchContext.Query != "" {
			var plan searchquery.Plan
			plan, err := searchquery.Pipeline(
				searchquery.Init(searchContext.Query, searchquery.SearchTypeRegex),
			)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to parse search query for search context: %s", rawContext)
			}
			inc, exc := plan.ToQ().Repositories()
			include = append(include, inc...)
			exclude = append(exclude, exc...)
		}
	}
	return include, exclude, nil
}

func (r *insightSeriesResolver) Status(ctx context.Context) (graphqlbackend.InsightStatusResolver, error) {
	seriesID := r.series.SeriesID

	status, err := queryrunner.QueryJobsStatus(ctx, r.workerBaseStore, seriesID)
	if err != nil {
		return nil, err
	}

	return NewStatusResolver(status, r.series.BackfillQueuedAt), nil
}

func (r *insightSeriesResolver) DirtyMetadata(ctx context.Context) ([]graphqlbackend.InsightDirtyQueryResolver, error) {
	data, err := r.metadataStore.GetDirtyQueriesAggregated(ctx, r.series.SeriesID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightDirtyQueryResolver, 0, len(data))
	for _, dqa := range data {
		resolvers = append(resolvers, &insightDirtyQueryResolver{dqa})
	}
	return resolvers, nil
}

var _ graphqlbackend.InsightsDataPointResolver = insightsDataPointResolver{}

type insightsDataPointResolver struct{ p store.SeriesPoint }

func (i insightsDataPointResolver) DateTime() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: i.p.Time}
}

func (i insightsDataPointResolver) Value() float64 { return i.p.Value }

type insightStatusResolver struct {
	totalPoints, pendingJobs, completedJobs, failedJobs int32
	backfillQueuedAt                                    *time.Time
}

func (i insightStatusResolver) TotalPoints() int32   { return i.totalPoints }
func (i insightStatusResolver) PendingJobs() int32   { return i.pendingJobs }
func (i insightStatusResolver) CompletedJobs() int32 { return i.completedJobs }
func (i insightStatusResolver) FailedJobs() int32    { return i.failedJobs }
func (i insightStatusResolver) BackfillQueuedAt() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(i.backfillQueuedAt)
}

func NewStatusResolver(status *queryrunner.JobsStatus, queuedAt *time.Time) *insightStatusResolver {
	return &insightStatusResolver{
		totalPoints: 0,

		// Include errored because they'll be retried before becoming failures
		pendingJobs: int32(status.Queued + status.Processing + status.Errored),

		completedJobs:    int32(status.Completed),
		failedJobs:       int32(status.Failed),
		backfillQueuedAt: queuedAt,
	}
}

type precalculatedInsightSeriesResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	series          types.InsightViewSeries
	metadataStore   store.InsightMetadataStore
	statusResolver  graphqlbackend.InsightStatusResolver

	seriesId string
	points   []store.SeriesPoint
	label    string
	filters  types.InsightViewFilters
}

func (p *precalculatedInsightSeriesResolver) SeriesId() string {
	return p.seriesId
}

func (p *precalculatedInsightSeriesResolver) Label() string {
	return p.label
}

func (p *precalculatedInsightSeriesResolver) Points(ctx context.Context, _ *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	resolvers := make([]graphqlbackend.InsightsDataPointResolver, 0, len(p.points))
	for _, point := range p.points {
		resolvers = append(resolvers, insightsDataPointResolver{point})
	}
	return resolvers, nil
}

func (p *precalculatedInsightSeriesResolver) Status(ctx context.Context) (graphqlbackend.InsightStatusResolver, error) {
	return p.statusResolver, nil
}

func (p *precalculatedInsightSeriesResolver) DirtyMetadata(ctx context.Context) ([]graphqlbackend.InsightDirtyQueryResolver, error) {
	data, err := p.metadataStore.GetDirtyQueriesAggregated(ctx, p.series.SeriesID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightDirtyQueryResolver, 0, len(data))
	for _, dqa := range data {
		resolvers = append(resolvers, &insightDirtyQueryResolver{dqa})
	}
	return resolvers, nil
}

type insightSeriesResolverGenerator interface {
	Generate(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error)
	handles(series types.InsightViewSeries) bool
	SetNext(nextGenerator insightSeriesResolverGenerator)
}

type handleSeriesFunc func(series types.InsightViewSeries) bool
type resolverGenerator func(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error)

type seriesResolverGenerator struct {
	next             insightSeriesResolverGenerator
	handlesSeries    handleSeriesFunc
	generateResolver resolverGenerator
}

func (j *seriesResolverGenerator) handles(series types.InsightViewSeries) bool {
	if j.handlesSeries == nil {
		return false
	}
	return j.handlesSeries(series)
}

func (j *seriesResolverGenerator) SetNext(nextGenerator insightSeriesResolverGenerator) {
	j.next = nextGenerator
}

func (j *seriesResolverGenerator) Generate(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error) {
	if j.handles(series) {
		return j.generateResolver(ctx, series, baseResolver, filters, seriesOptions)
	}
	if j.next != nil {
		return j.next.Generate(ctx, series, baseResolver, filters, seriesOptions)
	} else {
		log15.Error("no generator for insight series", "seriesID", series.SeriesID)
		return nil, errors.New("no resolvers for insights series")
	}
}

func newSeriesResolverGenerator(handles handleSeriesFunc, generate resolverGenerator) insightSeriesResolverGenerator {
	return &seriesResolverGenerator{
		handlesSeries:    handles,
		generateResolver: generate,
	}
}

func getRecordedSeriesPointOpts(ctx context.Context, db dbutil.DB, definition types.InsightViewSeries, filters types.InsightViewFilters) (*store.SeriesPointsOpts, error) {
	opts := &store.SeriesPointsOpts{}
	// Query data points only for the series we are representing.
	seriesID := definition.SeriesID
	opts.SeriesID = &seriesID

	// Default to last 12 points of data
	frames := query.BuildFrames(12, timeseries.TimeInterval{
		Unit:  types.IntervalUnit(definition.SampleIntervalUnit),
		Value: definition.SampleIntervalValue,
	}, time.Now())
	oldest := time.Now().AddDate(-1, 0, 0)
	if len(frames) != 0 {
		possibleOldest := frames[0].From
		if possibleOldest.Before(oldest) {
			oldest = possibleOldest
		}
	}
	opts.From = &oldest
	includeRepo := func(regex ...string) {
		opts.IncludeRepoRegex = append(opts.IncludeRepoRegex, regex...)
	}
	excludeRepo := func(regex ...string) {
		opts.ExcludeRepoRegex = append(opts.ExcludeRepoRegex, regex...)
	}

	if filters.IncludeRepoRegex != nil {
		includeRepo(*filters.IncludeRepoRegex)
	}
	if filters.ExcludeRepoRegex != nil {
		excludeRepo(*filters.ExcludeRepoRegex)
	}

	scLoader := &scLoader{primary: db}
	inc, exc, err := unwrapSearchContexts(ctx, scLoader, filters.SearchContexts)
	if err != nil {
		return nil, errors.Wrap(err, "unwrapSearchContexts")
	}
	includeRepo(inc...)
	excludeRepo(exc...)
	return opts, nil
}

func recordedSeries(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error) {
	opts, err := getRecordedSeriesPointOpts(ctx, r.workerBaseStore.Handle(), definition, filters)
	if err != nil {
		return nil, errors.Wrap(err, "getRecordedSeriesPointOpts")
	}

	points, err := r.timeSeriesStore.SeriesPoints(ctx, *opts)
	if err != nil {
		return nil, err
	}

	status, err := queryrunner.QueryJobsStatus(ctx, r.workerBaseStore, definition.SeriesID)
	if err != nil {
		return nil, errors.Wrap(err, "QueryJobsStatus")
	}
	statusResolver := NewStatusResolver(status, definition.BackfillQueuedAt)

	var resolvers []graphqlbackend.InsightSeriesResolver

	resolvers = append(resolvers, &precalculatedInsightSeriesResolver{
		insightsStore:   r.timeSeriesStore,
		workerBaseStore: r.workerBaseStore,
		series:          definition,
		metadataStore:   r.insightStore,
		points:          points,
		label:           definition.Label,
		filters:         filters,
		seriesId:        definition.SeriesID,
		statusResolver:  statusResolver,
	})
	return resolvers, nil
}

func expandCaptureGroupSeriesRecorded(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error) {
	opts, err := getRecordedSeriesPointOpts(ctx, r.workerBaseStore.Handle(), definition, filters)
	if err != nil {
		return nil, errors.Wrap(err, "getRecordedSeriesPointOpts")
	}

	groupedByCapture := make(map[string][]store.SeriesPoint)
	allPoints, err := r.timeSeriesStore.SeriesPoints(ctx, *opts)
	if err != nil {
		return nil, err
	}

	for i := range allPoints {
		point := allPoints[i]
		if point.Capture == nil {
			// skip nil values, this shouldn't be a real possibility
			continue
		}
		groupedByCapture[*point.Capture] = append(groupedByCapture[*point.Capture], point)
	}

	status, err := queryrunner.QueryJobsStatus(ctx, r.workerBaseStore, definition.SeriesID)
	if err != nil {
		return nil, errors.Wrap(err, "QueryJobsStatus")
	}
	statusResolver := NewStatusResolver(status, definition.BackfillQueuedAt)

	sortedCaptureGroups, limit := getSortedCaptureGroups(seriesOptions, definition, groupedByCapture)
	var resolvers []graphqlbackend.InsightSeriesResolver
	for _, capturedValue := range sortedCaptureGroups[0:limit] {
		points := groupedByCapture[capturedValue]
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time.Before(points[j].Time)
		})
		resolvers = append(resolvers, &precalculatedInsightSeriesResolver{
			insightsStore:   r.timeSeriesStore,
			workerBaseStore: r.workerBaseStore,
			series:          definition,
			metadataStore:   r.insightStore,
			points:          points,
			label:           capturedValue,
			filters:         filters,
			seriesId:        fmt.Sprintf("%s-%s", definition.SeriesID, capturedValue),
			statusResolver:  statusResolver,
		})
	}
	if len(resolvers) == 0 {
		// We are manually populating a mostly empty resolver here - this slightly hacky solution is to unify the
		// expectations of the webapp when querying for series state. For a standard search series there is
		// always a resolver since each series maps one to one with it's definition.
		// With a capture groups series we derive each unique series dynamically - which means it's possible to have a
		// series definition with zero resulting series. This most commonly occurs when the insight is just created,
		// before any data has been generated yet. Without this,
		// our capture groups insights don't share the loading state behavior.
		resolvers = append(resolvers, &precalculatedInsightSeriesResolver{
			insightsStore:   r.timeSeriesStore,
			workerBaseStore: r.workerBaseStore,
			series:          definition,
			metadataStore:   r.insightStore,
			statusResolver:  statusResolver,
			seriesId:        definition.SeriesID,
			points:          nil,
			label:           definition.Label,
			filters:         filters,
		})
	}
	return resolvers, nil
}

func expandCaptureGroupSeriesJustInTime(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error) {
	executor := query.NewCaptureGroupExecutor(r.postgresDB, time.Now)
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(definition.SampleIntervalUnit),
		Value: definition.SampleIntervalValue,
	}

	scLoader := &scLoader{primary: r.postgresDB}
	matchedRepos, err := filterRepositories(ctx, filters, definition.Repositories, scLoader)
	if err != nil {
		return nil, err
	}
	log15.Debug("capture group series", "seriesId", definition.SeriesID, "filteredRepos", matchedRepos)
	generatedSeries, err := executor.Execute(ctx, definition.Query, matchedRepos, interval)
	if err != nil {
		return nil, errors.Wrap(err, "CaptureGroupExecutor.Execute")
	}

	groupedByCapture := make(map[string][]store.SeriesPoint)
	var allPoints []store.SeriesPoint
	for i := range generatedSeries {
		for _, point := range generatedSeries[i].Points {
			allPoints = append(allPoints, store.SeriesPoint{
				SeriesID: generatedSeries[i].SeriesId,
				Time:     point.Time,
				Value:    float64(point.Count),
				Capture:  &generatedSeries[i].Label,
			})
		}
	}
	for i := range allPoints {
		point := allPoints[i]
		if point.Capture == nil {
			// skip nil values, this shouldn't be a real possibility
			continue
		}
		groupedByCapture[*point.Capture] = append(groupedByCapture[*point.Capture], point)
	}

	sortedCaptureGroups, limit := getSortedCaptureGroups(seriesOptions, definition, groupedByCapture)
	var resolvers []graphqlbackend.InsightSeriesResolver
	for _, capture := range sortedCaptureGroups[0:limit] {
		var series query.GeneratedTimeSeries
		for i := range generatedSeries {
			if generatedSeries[i].Label == capture {
				series = generatedSeries[i]
				break
			}
		}

		resolvers = append(resolvers, &dynamicInsightSeriesResolver{generated: &series})
	}

	return resolvers, nil
}

func streamingSeriesJustInTime(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters, seriesOptions types.SeriesDisplayOptions) ([]graphqlbackend.InsightSeriesResolver, error) {
	executor := query.NewStreamingExecutor(r.postgresDB, time.Now)
	interval := timeseries.TimeInterval{
		Unit:  types.IntervalUnit(definition.SampleIntervalUnit),
		Value: definition.SampleIntervalValue,
	}

	scLoader := &scLoader{primary: r.postgresDB}
	matchedRepos, err := filterRepositories(ctx, filters, definition.Repositories, scLoader)
	if err != nil {
		return nil, err
	}
	log15.Debug("just in time series", "seriesId", definition.SeriesID, "filteredRepos", matchedRepos)
	generatedSeries, err := executor.Execute(ctx, definition.Query, definition.Label, definition.SeriesID, matchedRepos, interval)
	if err != nil {
		return nil, errors.Wrap(err, "CaptureGroupExecutor.Execute")
	}

	var resolvers []graphqlbackend.InsightSeriesResolver
	for i := range generatedSeries {
		resolvers = append(resolvers, &dynamicInsightSeriesResolver{generated: &generatedSeries[i]})
	}

	return resolvers, nil
}

func getSortedCaptureGroups(seriesOptions types.SeriesDisplayOptions, definition types.InsightViewSeries, captureGroups map[string][]store.SeriesPoint) ([]string, int32) {
	var sortMode types.SeriesSortMode
	var sortDirection types.SeriesSortDirection
	var limit int32
	if seriesOptions.SortOptions != nil {
		sortMode = seriesOptions.SortOptions.Mode
		sortDirection = seriesOptions.SortOptions.Direction
	} else {
		if definition.SeriesSortMode != nil {
			sortMode = *definition.SeriesSortMode
		} else {
			sortMode = types.ResultCount
		}
		if definition.SeriesSortDirection != nil {
			sortDirection = *definition.SeriesSortDirection
		} else {
			sortDirection = types.Desc
		}
	}

	orderedCaptureGroups := make([][]store.SeriesPoint, 0, len(captureGroups))
	for _, value := range captureGroups {
		orderedCaptureGroups = append(orderedCaptureGroups, value)
	}

	getMostRecentValue := func(points []store.SeriesPoint) float64 {
		return points[len(points)-1].Value
	}

	ascLexSort := func(s1 string, s2 string) (hasSemVar bool, result bool) {
		version1, err1 := semver.NewVersion(s1)
		version2, err2 := semver.NewVersion(s2)
		if err1 == nil && err2 == nil {
			return true, version1.Compare(version2) < 0
		}
		if err1 != nil && err2 == nil {
			return true, false
		}
		if err1 == nil && err2 != nil {
			return true, true
		}
		return false, false
	}

	// First sort lexicographically (ascending) to make sure the ordering is consistent even if some result counts are equal.
	sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
		hasSemVar, result := ascLexSort(*orderedCaptureGroups[i][0].Capture, *orderedCaptureGroups[j][0].Capture)
		if hasSemVar == true {
			return result
		}
		return strings.Compare(*orderedCaptureGroups[i][0].Capture, *orderedCaptureGroups[j][0].Capture) < 0
	})

	switch sortMode {
	case types.ResultCount:
		if sortDirection == types.Asc {
			sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
				return getMostRecentValue(orderedCaptureGroups[i]) < getMostRecentValue(orderedCaptureGroups[j])
			})
		} else {
			sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
				return getMostRecentValue(orderedCaptureGroups[i]) > getMostRecentValue(orderedCaptureGroups[j])
			})
		}
	case types.Lexicographical:
		if sortDirection == types.Asc {
			// Already pre-sorted by default
		} else {
			sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
				hasSemVar, result := ascLexSort(*orderedCaptureGroups[i][0].Capture, *orderedCaptureGroups[j][0].Capture)
				if hasSemVar == true {
					return !result
				}
				return strings.Compare(*orderedCaptureGroups[i][0].Capture, *orderedCaptureGroups[j][0].Capture) > 0
			})
		}
	case types.DateAdded:
		if sortDirection == types.Asc {
			sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
				return orderedCaptureGroups[i][0].Time.Before(orderedCaptureGroups[j][0].Time)
			})
		} else {
			sort.SliceStable(orderedCaptureGroups, func(i, j int) bool {
				return orderedCaptureGroups[i][0].Time.After(orderedCaptureGroups[j][0].Time)
			})
		}
	}

	groupNames := []string{}
	for _, group := range orderedCaptureGroups {
		groupNames = append(groupNames, *group[0].Capture)
	}

	if seriesOptions.Limit != nil {
		limit = *seriesOptions.Limit
	} else if definition.SeriesLimit != nil {
		limit = *definition.SeriesLimit
	} else {
		limit = 20
	}
	limit = minInt(int(limit), len(groupNames))

	return groupNames, limit
}

func minInt(a, b int) int32 {
	if a < b {
		return int32(a)
	}
	return int32(b)
}
