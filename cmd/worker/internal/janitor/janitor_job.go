package janitor

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type janitorJob struct{}

func NewJanitorJob() job.Job {
	return &janitorJob{}
}

func (j *janitorJob) Description() string {
	return "Runs general background cleanup processes."
}

func (j *janitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *janitorJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	sqlDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, sqlDB)

	return []goroutine.BackgroundRoutine{
		// CheckRedisCacheEvictionPolicy(),
		// DeleteOldCacheDataInRedis(),
		// DeleteOldEventLogsInPostgres(context.Background(), db),
		// DeleteOldSecurityEventLogsInPostgres(context.Background(), db),
		// updatecheck.Start(db),
		newAnalyticsCacheRefreshRoutine(db),
	}, nil
}
