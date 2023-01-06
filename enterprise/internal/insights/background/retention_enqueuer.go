package background

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/retention"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newRetentionEnqueuer(ctx context.Context, observationCtx *observation.Context, workerBaseStore *basestore.Store, insightStore store.DataSeriesStore) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(ctx,
		"insights.retention.enqueuer",
		"enqueues series retention jobs",
		24*time.Hour,
		goroutine.HandlerFunc(
			func(ctx context.Context) error {
				// Only enqueue series with new data.
				seriesArgs := store.GetDataSeriesArgs{ExcludeJustInTime: true, NextRecordingBefore: time.Now()}
				allSeries, err := insightStore.GetDataSeries(ctx, seriesArgs)
				if err != nil {
					return errors.Wrap(err, "unable to fetch series for retention")
				}
				for _, series := range allSeries {
					_, err = retention.EnqueueJob(ctx, workerBaseStore, &retention.DataRetentionJob{SeriesID: series.ID})
					if err != nil {
						observationCtx.Logger.Error("could not enqueue data retention job", log.Int("seriesID", series.ID), log.Error(err))
					}
				}
				return nil
			}))
}
