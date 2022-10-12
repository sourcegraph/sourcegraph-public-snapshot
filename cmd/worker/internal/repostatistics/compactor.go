package repostatistics

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// compactor is a worker responsible for compacting rows in the repo_statistics table.
type compactor struct {
	observationContext *observation.Context
}

var _ job.Job = &compactor{}

func NewCompactor(observationContext *observation.Context) job.Job {
	return &compactor{observation.ContextWithLogger(log.NoOp(), observationContext)}
}

func (j *compactor) Description() string {
	return ""
}

func (j *compactor) Config() []env.Config {
	return nil
}

func (j *compactor) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(logger, j.observationContext)
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(context.Background(), 30*time.Minute, &handler{
			store:  db.RepoStatistics(),
			logger: logger,
		}),
	}, nil
}

type handler struct {
	store  database.RepoStatisticsStore
	logger log.Logger
}

var (
	_ goroutine.Handler      = &handler{}
	_ goroutine.ErrorHandler = &handler{}
)

func (h *handler) Handle(ctx context.Context) error {
	return h.store.CompactRepoStatistics(ctx)
}

func (h *handler) HandleError(err error) {
	h.logger.Error("error compacting repo statistics rows", log.Error(err))
}
