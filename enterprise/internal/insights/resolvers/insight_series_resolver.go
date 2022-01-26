package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/timeseries"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

var _ graphqlbackend.InsightSeriesResolver = &insightSeriesResolver{}

type insightSeriesResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	series          types.InsightViewSeries
	metadataStore   store.InsightMetadataStore

	filters types.InsightViewFilters
}

func (r *insightSeriesResolver) SeriesId() string { return r.series.SeriesID }

func (r *insightSeriesResolver) Label() string { return r.series.Label }

func (r *insightSeriesResolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var opts store.SeriesPointsOpts

	// Query data points only for the series we are representing.
	seriesID := r.series.SeriesID
	opts.SeriesID = &seriesID

	if args.From == nil {
		// Default to last 12mo of data
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
		args.From = &graphqlbackend.DateTime{Time: oldest}
	}
	if args.From != nil {
		opts.From = &args.From.Time
	}
	if args.To != nil {
		opts.To = &args.To.Time
	}

	// to preserve backwards compatibility, we are going to keep the arguments on this resolver for now. Ideally
	// we would deprecate these in favor of passing arguments from a higher level resolver (insight view) to match
	// the model of how we want default filters to work at the insight view level. That said, we will only inherit
	// higher resolver filters if provided filter arguments are nil.
	if args.IncludeRepoRegex != nil {
		opts.IncludeRepoRegex = *args.IncludeRepoRegex
	} else if r.filters.IncludeRepoRegex != nil {
		opts.IncludeRepoRegex = *r.filters.IncludeRepoRegex
	}
	if args.ExcludeRepoRegex != nil {
		opts.ExcludeRepoRegex = *args.ExcludeRepoRegex
	} else if r.filters.ExcludeRepoRegex != nil {
		opts.ExcludeRepoRegex = *r.filters.ExcludeRepoRegex
	}

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

func (r *insightSeriesResolver) Status(ctx context.Context) (graphqlbackend.InsightStatusResolver, error) {
	seriesID := r.series.SeriesID

	totalPoints, err := r.insightsStore.CountData(ctx, store.CountDataOpts{
		SeriesID: &seriesID,
	})
	if err != nil {
		return nil, err
	}

	status, err := queryrunner.QueryJobsStatus(ctx, r.workerBaseStore, seriesID)
	if err != nil {
		return nil, err
	}

	return insightStatusResolver{
		totalPoints: int32(totalPoints),

		// Include errored because they'll be retried before becoming failures
		pendingJobs: int32(status.Queued + status.Processing + status.Errored),

		completedJobs:    int32(status.Completed),
		failedJobs:       int32(status.Failed),
		backfillQueuedAt: r.series.BackfillQueuedAt,
	}, nil
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

type precalculatedInsightSeriesResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	series          types.InsightViewSeries
	metadataStore   store.InsightMetadataStore

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

func (p *precalculatedInsightSeriesResolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	resolvers := make([]graphqlbackend.InsightsDataPointResolver, 0, len(p.points))
	for _, point := range p.points {
		resolvers = append(resolvers, insightsDataPointResolver{point})
	}
	return resolvers, nil
}

func (p *precalculatedInsightSeriesResolver) Status(ctx context.Context) (graphqlbackend.InsightStatusResolver, error) {
	return insightStatusResolver{
		totalPoints:      0,
		pendingJobs:      0,
		completedJobs:    0,
		failedJobs:       0,
		backfillQueuedAt: p.series.BackfillQueuedAt,
	}, nil
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
