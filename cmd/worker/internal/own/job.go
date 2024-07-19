package own

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/own/background"
)

type ownRepoIndexingQueue struct{}

func NewOwnRepoIndexingQueue() job.Job {
	return &ownRepoIndexingQueue{}
}

func (o *ownRepoIndexingQueue) Description() string {
	return "Queue used to index ownership data partitioned per repository"
}

func (o *ownRepoIndexingQueue) Config() []env.Config {
	return nil
}

func (o *ownRepoIndexingQueue) Routines(startupCtx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !own.IsEnabled() {
		return nil, nil
	}
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	var routines []goroutine.BackgroundRoutine
	routines = append(routines, background.NewOwnBackgroundWorker(context.Background(), db, observationCtx)...)
	routines = append(routines, background.GetOwnIndexSchedulerRoutines(db, observationCtx)...)

	return routines, nil
}
