package search

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/uploadstore"
)

// config stores shared config we can override in each worker. We don't expose
// it as an env.Config since we currently only use it for testing.
type config struct {
	// WorkerInterval sets WorkerOptions.Interval for every worker
	WorkerInterval time.Duration
}

type searchJob struct {
	config config

	// workerDB if non-nil is used instead of calling workerdb.InitDB. Used
	// for testing
	workerDB database.DB

	once         sync.Once
	err          error
	workerStores []interface {
		QueuedCount(context.Context, bool) (int, error)
	}
	workers []goroutine.BackgroundRoutine
}

func NewSearchJob() job.Job {
	return &searchJob{
		config: config{
			WorkerInterval: 1 * time.Second,
		},
	}
}

func (j *searchJob) Description() string {
	return ""
}

func (j *searchJob) Config() []env.Config {
	return []env.Config{uploadstore.ConfigInst}
}

func (j *searchJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	workCtx := actor.WithInternalActor(context.Background())

	uploadStore, err := uploadstore.New(workCtx, observationCtx, uploadstore.ConfigInst)
	if err != nil {
		j.err = err
		return nil, err
	}

	newSearcherFactory := func(observationCtx *observation.Context, db database.DB) service.NewSearcher {
		searchClient := client.New(observationCtx.Logger, db, gitserver.NewClient("searchjobs.search"))
		return service.FromSearchClient(searchClient)
	}

	return j.newSearchJobRoutines(workCtx, observationCtx, uploadStore, newSearcherFactory)
}

func (j *searchJob) newSearchJobRoutines(
	workCtx context.Context,
	observationCtx *observation.Context,
	uploadStore uploadstore.Store,
	newSearcherFactory func(*observation.Context, database.DB) service.NewSearcher,
) ([]goroutine.BackgroundRoutine, error) {
	j.once.Do(func() {
		db := j.workerDB
		if db == nil {
			db, j.err = workerdb.InitDB(observationCtx)
			if j.err != nil {
				return
			}
		}

		newSearcher := newSearcherFactory(observationCtx, db)

		exhaustiveSearchStore := store.New(db, observationCtx)

		searchWorkerStore := store.NewExhaustiveSearchJobWorkerStore(observationCtx, db.Handle())
		repoWorkerStore := store.NewRepoSearchJobWorkerStore(observationCtx, db.Handle())
		revWorkerStore := store.NewRevSearchJobWorkerStore(observationCtx, db.Handle())

		j.workerStores = append(j.workerStores,
			searchWorkerStore,
			repoWorkerStore,
			revWorkerStore,
		)

		observationCtx = observation.ContextWithLogger(
			observationCtx.Logger.Scoped("routines"),
			observationCtx,
		)

		j.workers = []goroutine.BackgroundRoutine{
			newExhaustiveSearchWorker(workCtx, observationCtx, searchWorkerStore, exhaustiveSearchStore, newSearcher, j.config),
			newExhaustiveSearchRepoWorker(workCtx, observationCtx, repoWorkerStore, exhaustiveSearchStore, newSearcher, j.config),
			newExhaustiveSearchRepoRevisionWorker(workCtx, observationCtx, revWorkerStore, exhaustiveSearchStore, newSearcher, uploadStore, j.config),

			// resetters
			newExhaustiveSearchWorkerResetter(observationCtx, searchWorkerStore),
			newExhaustiveSearchRepoWorkerResetter(observationCtx, repoWorkerStore),
			newExhaustiveSearchRepoRevisionWorkerResetter(observationCtx, revWorkerStore),
		}
	})

	return j.workers, j.err
}

// hasWork returns true if any of the workers have work in its queue or is
// processing something. This is only exposed for tests.
func (j *searchJob) hasWork(ctx context.Context) bool {
	for _, w := range j.workerStores {
		if count, _ := w.QueuedCount(ctx, true); count > 0 {
			return true
		}
	}
	return false
}
