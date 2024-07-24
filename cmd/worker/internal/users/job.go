package users

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type aggregatedUsersStatisticsJob struct{}

func NewAggregatedUsersStatisticsJob() job.Job {
	return &aggregatedUsersStatisticsJob{}
}

func (e aggregatedUsersStatisticsJob) Description() string {
	return "updates the aggregated user statistics table in the database"
}

func (e aggregatedUsersStatisticsJob) Config() []env.Config {
	return nil
}

func (e aggregatedUsersStatisticsJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
			newAggregatedUsersStatisticsJob(observationCtx, db),
		},
		nil
}

func newAggregatedUsersStatisticsJob(observationCtx *observation.Context, db database.DB) goroutine.BackgroundRoutine {
	handler := goroutine.HandlerFunc(func(ctx context.Context) error {
		ff, err := db.FeatureFlags().GetFeatureFlag(ctx, "user_management_cache_disabled")
		if err == nil {
			disabled, _ := ff.EvaluateGlobal()
			if disabled {
				return nil
			}
		}
		return updateAggregatedUsersStatisticsTable(ctx, db)
	})

	operation := observationCtx.Operation(observation.Op{
		Name: "aggregated.user.statistics.update",
		Metrics: metrics.NewREDMetrics(
			observationCtx.Registerer,
			"update_aggregated_user_statistics",
			metrics.WithCountHelp("Total number of update_aggregated_user_statistics executions"),
		),
	})

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("update_aggregated_user_statistics"),
		goroutine.WithDescription("update aggregated user statistics in the database"),
		goroutine.WithInitialDelay(5*time.Minute),
		goroutine.WithInterval(12*time.Hour),
		goroutine.WithOperation(operation),
	)
}
