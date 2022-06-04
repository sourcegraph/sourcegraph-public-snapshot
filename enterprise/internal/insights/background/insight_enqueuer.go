package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/insights/priority"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// newInsightEnqueuer returns a background goroutine which will periodically find all of the search
// and webhook insights across all user settings, and enqueue work for the query runner and webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, workerBaseStore *basestore.Store, insightStore store.DataSeriesStore, featureFlagStore database.FeatureFlagStore, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"insights_enqueuer",
		metrics.WithCountHelp("Total number of insights enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "Enqueuer.Run",
		Metrics: metrics,
	})

	// Note: We run this goroutine once every hour, and StalledMaxAge in queryrunner/ is
	// set to 60s. If you change this, make sure the StalledMaxAge is less than this period
	// otherwise there is a fair chance we could enqueue work faster than it can be completed.
	//
	// See also https://github.com/sourcegraph/sourcegraph/pull/17227#issuecomment-779515187 for some very rough
	// data retention / scale concerns.
	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 1*time.Hour, goroutine.NewHandlerWithErrorMessage(
		"insights_enqueuer",
		func(ctx context.Context) error {
			queryRunnerEnqueueJob := func(ctx context.Context, job *queryrunner.Job) error {
				_, err := queryrunner.EnqueueJob(ctx, workerBaseStore, job)
				return err
			}
			now := time.Now

			return discoverAndEnqueueInsights(ctx, now, insightStore, featureFlagStore, queryRunnerEnqueueJob)
		},
	), operation)
}

func discoverAndEnqueueInsights(
	ctx context.Context,
	now func() time.Time,
	insightStore store.DataSeriesStore,
	ffs database.FeatureFlagStore,
	queryRunnerEnqueueJob func(ctx context.Context, job *queryrunner.Job) error) error {

	ctx = featureflag.WithFlags(ctx, ffs)
	flags := featureflag.FromContext(ctx)
	deprecateJustInTime := flags.GetBoolOr("code_insights_deprecate_jit", true)

	var multi error

	log15.Info("enqueuing indexed insight recordings")
	// this job will do the work of both recording (permanent) queries, and snapshot (ephemeral) queries. We want to try both, so if either has a soft-failure we will attempt both.
	recordingArgs := store.GetDataSeriesArgs{NextRecordingBefore: now(), ExcludeJustInTime: true}
	if !deprecateJustInTime {
		recordingArgs.GlobalOnly = true
	}
	recordingSeries, err := insightStore.GetDataSeries(ctx, recordingArgs)
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for recordings")
	}
	err = enqueue(ctx, recordingSeries, store.RecordMode, insightStore.StampRecording, queryRunnerEnqueueJob)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	log15.Info("enqueuing indexed insight snapshots")
	snapshotArgs := store.GetDataSeriesArgs{NextSnapshotBefore: now(), ExcludeJustInTime: true}
	if !deprecateJustInTime {
		snapshotArgs.GlobalOnly = true
	}
	snapshotSeries, err := insightStore.GetDataSeries(ctx, snapshotArgs)
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for snapshots")
	}
	err = enqueue(ctx, snapshotSeries, store.SnapshotMode, insightStore.StampSnapshot, queryRunnerEnqueueJob)
	if err != nil {
		multi = errors.Append(multi, err)
	}

	return multi
}

func enqueue(ctx context.Context, dataSeries []types.InsightSeries, mode store.PersistMode,
	stampFunc func(ctx context.Context, insightSeries types.InsightSeries) (types.InsightSeries, error),
	enqueueQueryRunnerJob func(ctx context.Context, job *queryrunner.Job) error) error {

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

		// Construct the search query that will generate data for this repository and time (revision) tuple.
		modifiedQuery, err := querybuilder.GlobalQuery(series.Query)
		if err != nil {
			multi = errors.Append(multi, errors.Wrapf(err, "GlobalQuery series_id:%s", seriesID))
			continue
		}

		err = enqueueQueryRunnerJob(ctx, &queryrunner.Job{
			SeriesID:    seriesID,
			SearchQuery: modifiedQuery,
			State:       "queued",
			Priority:    int(priority.High),
			Cost:        int(priority.Indexed),
			PersistMode: string(mode),
		})
		if err != nil {
			multi = errors.Append(multi, errors.Wrapf(err, "failed to enqueue insight series_id: %s", seriesID))
			continue
		}

		// The timestamp update can't be transactional because this is a separate database currently, so we will use
		// at-least-once semantics by waiting until the queue transaction is complete and without error.
		_, err = stampFunc(ctx, series)
		if err != nil {
			multi = errors.Append(multi, errors.Wrapf(err, "failed to stamp insight series_id: %s", seriesID))
			continue // might as well try the other insights and just skip this one
		}
		log15.Info("queued global search for insight recording", "series_id", series.SeriesID)
	}

	return multi
}
