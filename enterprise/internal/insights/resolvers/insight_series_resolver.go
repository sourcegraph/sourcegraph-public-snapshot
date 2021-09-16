package resolvers

import (
	"context"
	"time"

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
}

func (r *insightSeriesResolver) Label() string { return r.series.Label }

func (r *insightSeriesResolver) Points(ctx context.Context, args *graphqlbackend.InsightsPointsArgs) ([]graphqlbackend.InsightsDataPointResolver, error) {
	var opts store.SeriesPointsOpts

	// Query data points only for the series we are representing.
	seriesID := r.series.SeriesID
	opts.SeriesID = &seriesID

	if args.From == nil {
		// Default to last 12mo of data
		args.From = &graphqlbackend.DateTime{Time: time.Now().AddDate(-1, 0, 0)}
	}
	if args.From != nil {
		opts.From = &args.From.Time
	}
	if args.To != nil {
		opts.To = &args.To.Time
	}
	if args.IncludeRepoRegex != nil {
		opts.IncludeRepoRegex = *args.IncludeRepoRegex
	}
	if args.ExcludeRepoRegex != nil {
		opts.ExcludeRepoRegex = *args.ExcludeRepoRegex
	}
	// TODO(slimsag): future: Pass through opts.Limit

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
