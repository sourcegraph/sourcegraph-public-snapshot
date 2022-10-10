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
)

// compactor is a worker responsible for compacting rows in the repo_statistics table.
type compactor struct{}

var _ job.Job = &compactor{}

func NewCompactor() job.Job {
	return &compactor{}

}

func (j *compactor) Description() string {
	return ""
}

func (j *compactor) Config() []env.Config {
	return nil
}

func (j *compactor) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDBWithLogger(logger)
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

var _ goroutine.Handler = &handler{}
var _ goroutine.ErrorHandler = &handler{}

func (h *handler) Handle(ctx context.Context) error {
	return h.store.CompactRepoStatistics(ctx)
}

func (h *handler) HandleError(err error) {
	h.logger.Error("error compacting repo statistics rows", log.Error(err))
}
