package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
)

type searchJob struct{}

func NewSearchJob() job.Job {
	return &searchJob{}
}

func (j *searchJob) Description() string {
	return ""
}

func (j *searchJob) Config() []env.Config {
	return nil
}

func (j *searchJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	exhaustiveSearchStore := store.New(db, observationCtx)

	searchWorkerStore := store.NewExhaustiveSearchJobWorkerStore(observationCtx, db.Handle())
	repoWorkerStore := store.NewRepoSearchJobWorkerStore(observationCtx, db.Handle())
	revWorkerStore := store.NewRevSearchJobWorkerStore(observationCtx, db.Handle())

	observationCtx = observation.ContextWithLogger(
		observationCtx.Logger.Scoped("routines", "exhaustive search job routines"),
		observationCtx,
	)

	workCtx := actor.WithInternalActor(context.Background())
	return []goroutine.BackgroundRoutine{
		NewExhaustiveSearchWorker(workCtx, observationCtx, searchWorkerStore, exhaustiveSearchStore),
		NewExhaustiveSearchRepoWorker(workCtx, observationCtx, repoWorkerStore, exhaustiveSearchStore),
		NewExhaustiveSearchRepoRevisionWorker(workCtx, observationCtx, revWorkerStore, exhaustiveSearchStore),
	}, nil
}
