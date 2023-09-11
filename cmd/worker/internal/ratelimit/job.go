package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type rateLimitConfigJob struct{}

func NewRateLimitConfigJob() job.Job {
	return &rateLimitConfigJob{}
}

func (s *rateLimitConfigJob) Description() string {
	return "Copies the rate limit configurations from the database to Redis."
}

func (s *rateLimitConfigJob) Config() []env.Config {
	return nil
}

func (s *rateLimitConfigJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	//TODO: Allow this job to run once an in memory version is available
	if deploy.IsApp() {
		return nil, nil
	}
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	rl, err := redispool.NewRateLimiter()
	if err != nil {
		return nil, err
	}
	rlcWorker := makeRateLimitConfigWorker(&handler{
		logger:        observationCtx.Logger.Scoped("Periodic rate limit config job", "Routine that periodically copies rate limit configurations for code hosts from the database to Redis."),
		codeHostStore: db.CodeHosts(),
		rateLimiter:   ratelimit.NewCodeHostRateLimiter(rl),
	})

	return []goroutine.BackgroundRoutine{rlcWorker}, nil
}

func makeRateLimitConfigWorker(handler *handler) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		handler,
		goroutine.WithName("rate_limit_config_worker"),
		goroutine.WithDescription("copies the rate limit configurations from the database to Redis"),
		goroutine.WithInterval(30*time.Second),
	)
}
