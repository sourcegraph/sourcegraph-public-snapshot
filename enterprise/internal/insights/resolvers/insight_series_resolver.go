package resolvers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"

	"github.com/sourcegraph/sourcegraph/internal/metrics"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	sctypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.InsightSeriesResolver = &precalculatedInsightSeriesResolver{}

// SearchContextLoader loads search contexts just from the full name of the
// context. This will not verify that the calling context owns the context, it
// will load regardless of the current user.
type SearchContextLoader interface {
	GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error)
}

type scLoader struct {
	primary database.DB
}

func (l *scLoader) GetByName(ctx context.Context, name string) (*sctypes.SearchContext, error) {
	return searchcontexts.ResolveSearchContextSpec(ctx, l.primary, name)
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

var _ graphqlbackend.InsightsDataPointResolver = insightsDataPointResolver{}

type insightsDataPointResolver struct{ p store.SeriesPoint }

func (i insightsDataPointResolver) DateTime() gqlutil.DateTime {
	return gqlutil.DateTime{Time: i.p.Time}
}

func (i insightsDataPointResolver) Value() float64 { return i.p.Value }

type statusInfo struct {
	totalPoints, pendingJobs, completedJobs, failedJobs int32
	backfillQueuedAt                                    *time.Time
	isLoading                                           bool
	incompletedDatapoints                               []store.IncompleteDatapoint
}

type GetSeriesQueueStatusFunc func(ctx context.Context, seriesID string) (*queryrunner.JobsStatus, error)
type GetSeriesBackfillsFunc func(ctx context.Context, seriesID int) ([]scheduler.SeriesBackfill, error)
type GetIncompleteDatapointsFunc func(ctx context.Context, seriesID int) ([]store.IncompleteDatapoint, error)
type insightStatusResolver struct {
	getQueueStatus          GetSeriesQueueStatusFunc
	getSeriesBackfills      GetSeriesBackfillsFunc
	getIncompleteDatapoints GetIncompleteDatapointsFunc
	statusOnce              sync.Once
	series                  types.InsightViewSeries

	status    statusInfo
	statusErr error
}

func (i *insightStatusResolver) TotalPoints(ctx context.Context) (int32, error) {
	status, err := i.calculateStatus(ctx)
	return status.totalPoints, err
}
func (i *insightStatusResolver) PendingJobs(ctx context.Context) (int32, error) {
	status, err := i.calculateStatus(ctx)
	return status.pendingJobs, err
}
func (i *insightStatusResolver) CompletedJobs(ctx context.Context) (int32, error) {
	status, err := i.calculateStatus(ctx)
	return status.completedJobs, err
}
func (i *insightStatusResolver) FailedJobs(ctx context.Context) (int32, error) {
	status, err := i.calculateStatus(ctx)
	return status.failedJobs, err
}
func (i *insightStatusResolver) BackfillQueuedAt(ctx context.Context) *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(i.series.BackfillQueuedAt)
}
func (i *insightStatusResolver) IsLoadingData(ctx context.Context) (*bool, error) {
	status, err := i.calculateStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &status.isLoading, nil
}

func (i *insightStatusResolver) calculateStatus(ctx context.Context) (statusInfo, error) {
	i.statusOnce.Do(func() {
		status, statusErr := i.getQueueStatus(ctx, i.series.SeriesID)
		if statusErr != nil {
			i.statusErr = errors.Wrap(statusErr, "QueryJobsStatus")
			return
		}
		i.status.backfillQueuedAt = i.series.BackfillQueuedAt
		i.status.completedJobs = int32(status.Completed)
		i.status.failedJobs = int32(status.Failed)
		i.status.pendingJobs = int32(status.Queued + status.Processing + status.Errored)

		seriesBackfills, backillErr := i.getSeriesBackfills(ctx, i.series.InsightSeriesID)
		if backillErr != nil {
			i.statusErr = errors.Wrap(backillErr, "LoadSeriesBackfills")
			return
		}
		backfillInProgress := false
		for n := range seriesBackfills {
			if seriesBackfills[n].SeriesId == i.series.InsightSeriesID && !seriesBackfills[n].IsTerminalState() {
				backfillInProgress = true
				break
			}
		}
		i.status.isLoading = i.status.backfillQueuedAt == nil || i.status.pendingJobs > 0 || backfillInProgress
	})
	return i.status, i.statusErr
}

func NewStatusResolver(r *baseInsightResolver, viewSeries types.InsightViewSeries) *insightStatusResolver {
	getStatus := func(ctx context.Context, series string) (*queryrunner.JobsStatus, error) {
		return queryrunner.QueryJobsStatus(ctx, r.workerBaseStore, series)
	}
	getBackfills := func(ctx context.Context, seriesID int) ([]scheduler.SeriesBackfill, error) {
		backfillStore := scheduler.NewBackfillStore(r.insightsDB)
		return backfillStore.LoadSeriesBackfills(ctx, seriesID)
	}
	getIncompletes := func(ctx context.Context, seriesID int) ([]store.IncompleteDatapoint, error) {
		return r.timeSeriesStore.LoadAggregatedIncompleteDatapoints(ctx, seriesID)
	}
	return newStatusResolver(getStatus, getBackfills, getIncompletes, viewSeries)
}

func newStatusResolver(getQueueStatus GetSeriesQueueStatusFunc, getSeriesBackfills GetSeriesBackfillsFunc, getIncompleteDatapoints GetIncompleteDatapointsFunc, series types.InsightViewSeries) *insightStatusResolver {
	return &insightStatusResolver{
		getQueueStatus:          getQueueStatus,
		getSeriesBackfills:      getSeriesBackfills,
		series:                  series,
		getIncompleteDatapoints: getIncompleteDatapoints,
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
	modifiedPoints := removeClosePoints(p.points, p.series)
	for _, point := range modifiedPoints {
		resolvers = append(resolvers, insightsDataPointResolver{point})
	}
	return resolvers, nil
}

// This will make sure that no two snapshots are too close together. We'll use 20% of the time interval to
// remove these "close" points.
func removeClosePoints(points []store.SeriesPoint, series types.InsightViewSeries) []store.SeriesPoint {
	buffer := intervalToMinutes(types.IntervalUnit(series.SampleIntervalUnit), series.SampleIntervalValue) / 5
	modifiedPoints := []store.SeriesPoint{}
	for i := 0; i < len(points)-1; i++ {
		modifiedPoints = append(modifiedPoints, points[i])
		if points[i+1].Time.Sub(points[i].Time).Minutes() < buffer {
			i++
		}
	}
	// Always add the very last snapshot point if it exists
	if len(points) > 0 {
		return append(modifiedPoints, points[len(points)-1])
	}
	return modifiedPoints
}

// This only needs to be approximate to calculate a comfortable buffer in which to remove points
func intervalToMinutes(unit types.IntervalUnit, value int) float64 {
	switch unit {
	case types.Day:
		return time.Hour.Minutes() * 24 * float64(value)
	case types.Week:
		return time.Hour.Minutes() * 24 * 7 * float64(value)
	case types.Month:
		return time.Hour.Minutes() * 24 * 30 * float64(value)
	case types.Year:
		return time.Hour.Minutes() * 24 * 365 * float64(value)
	default:
		// By default return the smallest interval (an hour)
		return time.Hour.Minutes() * float64(value)
	}
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
	Generate(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error)
	handles(series types.InsightViewSeries) bool
	SetNext(nextGenerator insightSeriesResolverGenerator)
}

type handleSeriesFunc func(series types.InsightViewSeries) bool
type resolverGenerator func(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error)

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

func (j *seriesResolverGenerator) Generate(ctx context.Context, series types.InsightViewSeries, baseResolver baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error) {
	if j.handles(series) {
		return j.generateResolver(ctx, series, baseResolver, filters)
	}
	if j.next != nil {
		return j.next.Generate(ctx, series, baseResolver, filters)
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

func getRecordedSeriesPointOpts(ctx context.Context, db database.DB, definition types.InsightViewSeries, filters types.InsightViewFilters) (*store.SeriesPointsOpts, error) {
	opts := &store.SeriesPointsOpts{}
	// Query data points only for the series we are representing.
	seriesID := definition.SeriesID
	opts.SeriesID = &seriesID
	opts.ID = &definition.InsightSeriesID
	opts.SupportsAugmentation = definition.SupportsAugmentation

	// Default to last 12 points of data
	frames := timeseries.BuildFrames(12, timeseries.TimeInterval{
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

var loadingStrategyRED = metrics.NewREDMetrics(prometheus.DefaultRegisterer, "src_insights_loading_strategy", metrics.WithLabels("in_mem", "capture"))

func fetchSeries(ctx context.Context, definition types.InsightViewSeries, filters types.InsightViewFilters, r *baseInsightResolver) (points []store.SeriesPoint, err error) {
	opts, err := getRecordedSeriesPointOpts(ctx, database.NewDBWith(log.Scoped("recordedSeries", ""), r.workerBaseStore), definition, filters)
	if err != nil {
		return nil, errors.Wrap(err, "getRecordedSeriesPointOpts")
	}

	getAltFlag := func() bool {
		ex := conf.Get().ExperimentalFeatures
		if ex == nil {
			return false
		}
		return ex.InsightsAlternateLoadingStrategy
	}
	alternativeLoadingStrategy := getAltFlag()

	var start, end time.Time
	start = time.Now()
	if !alternativeLoadingStrategy {
		points, err = r.timeSeriesStore.SeriesPoints(ctx, *opts)
		if err != nil {
			return nil, err
		}
	} else {
		points, err = r.timeSeriesStore.LoadSeriesInMem(ctx, *opts)
		if err != nil {
			return nil, err
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time.Before(points[j].Time)
		})
	}
	end = time.Now()
	loadingStrategyRED.Observe(end.Sub(start).Seconds(), 1, &err, strconv.FormatBool(alternativeLoadingStrategy), strconv.FormatBool(definition.GeneratedFromCaptureGroups))

	return points, err
}

func recordedSeries(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters) (_ []graphqlbackend.InsightSeriesResolver, err error) {
	points, err := fetchSeries(ctx, definition, filters, &r)
	if err != nil {
		return nil, err
	}

	statusResolver := NewStatusResolver(&r, definition)

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

func expandCaptureGroupSeriesRecorded(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error) {
	allPoints, err := fetchSeries(ctx, definition, filters, &r)
	if err != nil {
		return nil, err
	}
	groupedByCapture := make(map[string][]store.SeriesPoint)

	for i := range allPoints {
		point := allPoints[i]
		if point.Capture == nil {
			// skip nil values, this shouldn't be a real possibility
			continue
		}
		groupedByCapture[*point.Capture] = append(groupedByCapture[*point.Capture], point)
	}

	statusResolver := NewStatusResolver(&r, definition)

	var resolvers []graphqlbackend.InsightSeriesResolver
	for capturedValue, points := range groupedByCapture {
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

func expandCaptureGroupSeriesJustInTime(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error) {
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

	var resolvers []graphqlbackend.InsightSeriesResolver
	for i := range generatedSeries {
		resolvers = append(resolvers, &dynamicInsightSeriesResolver{generated: &generatedSeries[i]})
	}

	return resolvers, nil
}

func streamingSeriesJustInTime(ctx context.Context, definition types.InsightViewSeries, r baseInsightResolver, filters types.InsightViewFilters) ([]graphqlbackend.InsightSeriesResolver, error) {
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
		return nil, errors.Wrap(err, "StreamingQueryExecutor.Execute")
	}

	var resolvers []graphqlbackend.InsightSeriesResolver
	for i := range generatedSeries {
		resolvers = append(resolvers, &dynamicInsightSeriesResolver{generated: &generatedSeries[i]})
	}

	return resolvers, nil
}

var _ graphqlbackend.TimeoutDatapointAlert = &timeoutDatapointAlertResolver{}

// var _ graphqlbackend.IncompleteDatapointAlert = &timeoutDatapointAlertResolver{}
var _ graphqlbackend.IncompleteDatapointAlert = &IncompleteDataPointAlertResolver{}

type IncompleteDataPointAlertResolver struct {
	// resolver any
	// graphqlbackend.IncompleteDatapointAlert
	point store.IncompleteDatapoint
}

func (i *IncompleteDataPointAlertResolver) ToTimeoutDatapointAlert() (graphqlbackend.TimeoutDatapointAlert, bool) {
	if i.point.Reason == store.ReasonTimeout {
		return &timeoutDatapointAlertResolver{point: i.point}, true
	}

	// t, ok := i.resolver.(graphqlbackend.TimeoutDatapointAlert)
	// return t, ok
	return nil, false
}

func (i *IncompleteDataPointAlertResolver) Time() gqlutil.DateTime {
	return gqlutil.DateTime{Time: i.point.Time}
}

type timeoutDatapointAlertResolver struct {
	point store.IncompleteDatapoint
	baseInsightResolver
}

func (t *timeoutDatapointAlertResolver) Time() gqlutil.DateTime {
	return gqlutil.DateTime{Time: t.point.Time}
}

func (i *insightStatusResolver) IncompleteDatapoints(ctx context.Context) (resolvers []graphqlbackend.IncompleteDatapointAlert, err error) {
	incomplete, err := i.getIncompleteDatapoints(ctx, i.series.InsightSeriesID)
	for _, reason := range incomplete {
		resolvers = append(resolvers, &IncompleteDataPointAlertResolver{point: reason})
	}

	return resolvers, err
}
