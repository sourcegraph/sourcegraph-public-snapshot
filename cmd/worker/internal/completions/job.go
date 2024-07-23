package completions

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type tokenUsageJob struct{}

func NewTokenUsageJob() job.Job {
	return &tokenUsageJob{}
}

func (e tokenUsageJob) Description() string {
	return "stores LLM token usage in DB"
}

func (e tokenUsageJob) Config() []env.Config {
	return nil
}

func (e tokenUsageJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
			newTokenUsageJob(observationCtx, db),
		},
		nil
}

func newTokenUsageJob(observationCtx *observation.Context, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		return recordTokenUsage(ctx, tokenusage.NewManager(), telemetryrecorder.New(db))
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "cody.llmTokenCounter.record",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"cody_llm_token_counter",
			metrics.WithCountHelp("Total number of cody_llm_token_counter executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("cody_llm_token_counter"),
		goroutine.WithDescription("Stores LLM token usage in DB"),
		goroutine.WithInterval(5*time.Minute),
		goroutine.WithOperation(operation),
	)
}
