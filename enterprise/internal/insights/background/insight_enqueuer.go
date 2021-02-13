package background

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// newInsightEnqueuer returns a background goroutine which will periodically find all of the search
// and webhook insights across all user settings, and enqueue work for the query runner and webhook
// runner workers to perform.
func newInsightEnqueuer(ctx context.Context, workerBaseStore *basestore.Store, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"insights_enqueuer",
		metrics.WithCountHelp("Total number of insights enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    fmt.Sprintf("Enqueuer.Run"),
		Metrics: metrics,
	})

	// TODO(slimsag): future: before deploying to prod, confirm retention policy is OK with 1 minute
	// intervals / consider if we need to adjust the interval.

	// Note: We run this goroutine once very 1 minute, and StalledMaxAge in queryrunner/ is
	// set to 60s. If you change this, make sure the StalledMaxAge is less than this period
	// otherwise there is a fair chance we could enqueue work faster than it can be completed.
	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 1*time.Minute, goroutine.NewHandlerWithErrorMessage(
		"insights_enqueuer",
		func(ctx context.Context) error {
			// TODO(slimsag): future: discover insights from settings store and enqueue them here.
			// _, err := queryrunner.EnqueueJob(ctx, workerBaseStore, &queryrunner.Job{
			// 	SeriesID:    "abcdefg", // TODO(slimsag)
			// 	SearchQuery: "errorf",  // TODO(slimsag)
			// 	State:       "queued",
			// })
			return nil
		},
	), operation)
}
