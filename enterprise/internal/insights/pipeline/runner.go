package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrSeriesNotBackfilling error = errors.New("series not backfilling")

func NewPipelineRunner(ctx context.Context, dataSeriesStore store.DataSeriesStore, pipelineFactory BackfillerFactory, observationContext *observation.Context) goroutine.BackgroundRoutine {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"insights_historical_enqueuer",
		metrics.WithCountHelp("Total number of insights historical enqueuer executions"),
	)
	operation := observationContext.Operation(observation.Op{
		Name:    "InsightBackfillPipeline.Run",
		Metrics: metrics,
	})

	runner := &pipelineRunner{
		dataSeriesStore:  dataSeriesStore,
		pipelieneFactory: pipelineFactory,
		logger:           log.Scoped("insights_pipeline_runner", ""),
		runningSeries:    map[string]BackfillProgress{},
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(ctx, 1*time.Minute, goroutine.NewHandlerWithErrorMessage(
		"insights_historical_pipeline",
		runner.Handler,
	), operation)

}

type pipelineRunner struct {
	dataSeriesStore  store.DataSeriesStore
	pipelieneFactory BackfillerFactory
	logger           log.Logger
	runningSeries    map[string]BackfillProgress
	mu               sync.RWMutex
}

func (h *pipelineRunner) updateProgress(progress BackfillProgress) {
	h.logger.Warn("progress update", log.String("seriesID", progress.SeriesID), log.Intp("remaining cost", progress.RemaingCost))
	h.mu.Lock()
	h.runningSeries[progress.SeriesID] = progress
	h.mu.Unlock()
}

func (h *pipelineRunner) getSeriesProgress(seriesID string) (*BackfillProgress, error) {
	h.mu.RLock()
	p, ok := h.runningSeries[seriesID]
	if !ok {
		return nil, ErrSeriesNotBackfilling
	}
	return &p, nil
}
func (h *pipelineRunner) Handler(ctx context.Context) error {
	// 🚨 SECURITY: This background process uses the internal actor to interact with Sourcegraph services. This background process
	// is responsible for calculating the work needed to backfill an insight series _without_ a user context. Repository permissions
	// are filtered at view time of an insight.
	ctx = actor.WithInternalActor(ctx)

	//TODO: Add a first step here that creates a backfill for new series and "scores" the cost of it

	//TODO: Update this to pull all the not complete backfills
	foundInsights, err := h.dataSeriesStore.GetDataSeries(ctx, store.GetDataSeriesArgs{BackfillNotQueued: true, ExcludeJustInTime: true})
	if err != nil {
		return errors.Wrap(err, "Discover")
	}

	var multi error
	for _, series := range foundInsights {
		incrementErr := h.dataSeriesStore.IncrementBackfillAttempts(ctx, series)
		if incrementErr != nil {
			h.logger.Warn("unabled to increment backfill attempts", log.String("series", series.SeriesID))
		}
		seriesPipeline := h.pipelieneFactory(&series)
		err := seriesPipeline.Run(ctx, h.updateProgress)
		if err != nil {
			h.logger.Warn("backfill failed", log.String("series", series.SeriesID))
		}
		multi = errors.Append(err, multi)

	}

	return multi
}
