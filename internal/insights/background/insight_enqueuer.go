package background

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newInsightEnqueuer returns a background goroutine which will periodically find all of the search
// and webhook insights across all user settings, and enqueue work for the query runner and webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, observationCtx *observation.Context, workerBaseStore *basestore.Store, insightStore store.DataSeriesStore, logger log.Logger) goroutine.BackgroundRoutine {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"insights_enqueuer",
		metrics.WithCountHelp("Total number of insights enqueuer executions"),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "Enqueuer.Run",
		Metrics: redMetrics,
	})

	// Note: We run this goroutine once every hour, and StalledMaxAge in queryrunner/ is
	// set to 60s. If you change this, make sure the StalledMaxAge is less than this period
	// otherwise there is a fair chance we could enqueue work faster than it can be completed.
	//
	// See also https://github.com/sourcegraph/sourcegraph/pull/17227#issuecomment-779515187 for some very rough
	// data retention / scale concerns.
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(
			func(ctx context.Context) error {
				ie := NewInsightEnqueuer(time.Now, workerBaseStore, logger)

				return ie.discoverAndEnqueueInsights(ctx, insightStore)
			},
		),
		goroutine.WithName("insights.enqueuer"),
		goroutine.WithDescription("enqueues snapshot and current recording query jobs"),
		goroutine.WithInterval(1*time.Hour),
		goroutine.WithOperation(operation),
	)
}

type InsightEnqueuer struct {
	logger log.Logger

	now                   func() time.Time
	enqueueQueryRunnerJob func(context.Context, *queryrunner.Job) error
}

func NewInsightEnqueuer(now func() time.Time, workerBaseStore *basestore.Store, logger log.Logger) *InsightEnqueuer {
	return &InsightEnqueuer{
		now: now,
		enqueueQueryRunnerJob: func(ctx context.Context, job *queryrunner.Job) error {
			_, err := queryrunner.EnqueueJob(ctx, workerBaseStore, job)
			return err
		},
		logger: logger,
	}
}

func (ie *InsightEnqueuer) discoverAndEnqueueInsights(
	ctx context.Context,
	insightStore store.DataSeriesStore,
) error {
	var multi error

	ie.logger.Info("enqueuing indexed insight recordings")
	// this job will do the work of both recording (permanent) queries, and snapshot (ephemeral) queries. We want to try both, so if either has a soft-failure we will attempt both.
	recordingArgs := store.GetDataSeriesArgs{NextRecordingBefore: ie.now(), ExcludeJustInTime: true}
	recordingSeries, err := insightStore.GetDataSeries(ctx, recordingArgs)
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for recordings")
	}
	err = ie.Enqueue(ctx, recordingSeries, store.RecordMode, insightStore.StampRecording)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	ie.logger.Info("enqueuing indexed insight snapshots")
	snapshotArgs := store.GetDataSeriesArgs{NextSnapshotBefore: ie.now(), ExcludeJustInTime: true}
	snapshotSeries, err := insightStore.GetDataSeries(ctx, snapshotArgs)
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for snapshots")
	}
	err = ie.Enqueue(ctx, snapshotSeries, store.SnapshotMode, insightStore.StampSnapshot)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	return multi
}

func (ie *InsightEnqueuer) Enqueue(
	ctx context.Context,
	dataSeries []types.InsightSeries,
	mode store.PersistMode,
	stampFunc func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error),
) error {
	// Deduplicate series that may be unique (e.g. different name/description) but do not have
	// unique data (i.e. use the same exact search query or webhook URL.)
	var (
		uniqueSeries = map[string]types.InsightSeries{}
		multi        error
	)
	for _, series := range dataSeries {
		seriesID := series.SeriesID
		_, enqueuedAlready := uniqueSeries[seriesID]
		if enqueuedAlready {
			continue
		}
		uniqueSeries[seriesID] = series

		if err := ie.EnqueueSingle(ctx, series, mode, stampFunc); err != nil {
			multi = errors.Append(multi, err)
		}
	}

	return multi
}

func (ie *InsightEnqueuer) EnqueueSingle(
	ctx context.Context,
	series types.InsightSeries,
	mode store.PersistMode,
	stampFunc func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error),
) error {
	// Construct the search query that will generate data for this repository and time (revision) tuple.
	defaultQueryParams := querybuilder.CodeInsightsQueryDefaults(len(series.Repositories) == 0)
	seriesID := series.SeriesID
	var err error

	basicQuery := querybuilder.BasicQuery(series.Query)
	var modifiedQuery querybuilder.BasicQuery
	var finalQuery string

	if series.RepositoryCriteria != nil {
		modifiedQuery, err = querybuilder.MakeQueryWithRepoFilters(*series.RepositoryCriteria, basicQuery, true, querybuilder.CodeInsightsQueryDefaults(true)...)
	} else if len(series.Repositories) > 0 {
		modifiedQuery, err = querybuilder.MultiRepoQuery(basicQuery, series.Repositories, defaultQueryParams)
	} else {
		modifiedQuery, err = querybuilder.GlobalQuery(basicQuery, defaultQueryParams)
	}
	if err != nil {
		return errors.Wrapf(err, "GlobalQuery series_id:%s", seriesID)
	}
	finalQuery = modifiedQuery.String()
	if series.GroupBy != nil {
		computeQuery, err := querybuilder.ComputeInsightCommandQuery(modifiedQuery, querybuilder.MapType(*series.GroupBy), gitserver.NewClient("insights.enqueuer"))
		if err != nil {
			return errors.Wrapf(err, "ComputeInsightCommandQuery series_id:%s", seriesID)
		}
		finalQuery = computeQuery.String()
	}

	err = ie.enqueueQueryRunnerJob(ctx, &queryrunner.Job{
		SearchJob: queryrunner.SearchJob{
			SeriesID:    seriesID,
			SearchQuery: finalQuery,
			PersistMode: string(mode),
		},
		State:    "queued",
		Priority: int(priority.High),
		Cost:     int(priority.Indexed),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to enqueue insight series_id: %s", seriesID)
	}

	// The timestamp update can't be transactional because this is a separate database currently, so we will use
	// at-least-once semantics by waiting until the queue transaction is complete and without error.
	_, err = stampFunc(ctx, series)
	if err != nil {
		// might as well try the other insights and just skip this one
		return errors.Wrapf(err, "failed to stamp insight series_id: %s", seriesID)
	}

	ie.logger.Info("queued global search for insight", log.String("persist mode", string(mode)), log.String("seriesID", series.SeriesID))
	return nil
}
