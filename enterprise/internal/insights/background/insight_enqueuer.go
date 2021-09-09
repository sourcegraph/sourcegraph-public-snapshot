package background

import (
	"context"
	"strings"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/insights/priority"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// newInsightEnqueuer returns a background goroutine which will periodically find all of the search
// and webhook insights across all user settings, and enqueue work for the query runner and webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, workerBaseStore *basestore.Store, insightStore store.DataSeriesStore, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"insights_enqueuer",
		metrics.WithCountHelp("Total number of insights enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "Enqueuer.Run",
		Metrics: metrics,
	})

	// Note: We run this goroutine once every 10 minutes, and StalledMaxAge in queryrunner/ is
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

			return discoverAndEnqueueInsights(ctx, now, insightStore, queryRunnerEnqueueJob)
		},
	), operation)
}

func discoverAndEnqueueInsights(
	ctx context.Context,
	now func() time.Time,
	insightStore store.DataSeriesStore,
	queryRunnerEnqueueJob func(ctx context.Context, job *queryrunner.Job) error) error {

	var multi error

	log15.Info("enqueuing indexed insight recordings")
	// this job will do the work of both recording (permanent) queries, and snapshot (ephemeral) queries. We want to try both, so if either has a soft-failure we will attempt both.
	recordingSeries, err := insightStore.GetDataSeries(ctx, store.GetDataSeriesArgs{NextRecordingBefore: now()})
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for recordings")
	}
	err = enqueue(ctx, recordingSeries, store.RecordMode, insightStore.StampRecording, queryRunnerEnqueueJob)
	if err != nil {
		multi = multierror.Append(multi, err)
	}

	log15.Info("enqueuing indexed insight snapshots")
	snapshotSeries, err := insightStore.GetDataSeries(ctx, store.GetDataSeriesArgs{NextSnapshotBefore: now()})
	if err != nil {
		return errors.Wrap(err, "indexed insight recorder: unable to fetch series for snapshots")
	}
	err = enqueue(ctx, snapshotSeries, store.SnapshotMode, insightStore.StampSnapshot, queryRunnerEnqueueJob)
	if err != nil {
		multi = multierror.Append(multi, err)
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

		err := enqueueQueryRunnerJob(ctx, &queryrunner.Job{
			SeriesID:    seriesID,
			SearchQuery: withCountUnlimited(series.Query),
			State:       "queued",
			Priority:    int(priority.High),
			Cost:        int(priority.Indexed),
			PersistMode: string(mode),
		})
		if err != nil {
			multi = multierror.Append(multi, errors.Wrapf(err, "failed to enqueue insight series_id: %s", seriesID))
			continue
		}

		// The timestamp update can't be transactional because this is a separate database currently, so we will use
		// at-least-once semantics by waiting until the queue transaction is complete and without error.
		_, err = stampFunc(ctx, series)
		if err != nil {
			multi = multierror.Append(multi, errors.Wrapf(err, "failed to stamp insight series_id: %s", seriesID))
			continue // might as well try the other insights and just skip this one
		}
	}

	return multi
}

// withCountUnlimited adds `count:9999999` to the given search query string iff `count:` does not
// exist in the query string. This is extremely important as otherwise the number of results our
// search query would return would be incomplete and fluctuate.
//
// TODO(slimsag): future: we should pull in the search query parser to avoid cases where `count:`
// is actually e.g. a search query like `content:"count:"`.
func withCountUnlimited(s string) string {
	if strings.Contains(s, "count:") {
		return s
	}
	return s + " count:all"
}
