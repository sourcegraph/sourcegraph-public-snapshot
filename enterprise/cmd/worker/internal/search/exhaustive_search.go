package search

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewExhaustiveSearchWorker creates a background routine that periodically runs the exhaustive search.
func NewExhaustiveSearchWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*exhaustive.Job],
) goroutine.BackgroundRoutine {
	handler := &exhaustiveSearch{
		logger: log.Scoped("exhaustive-search", "The background worker running exhaustive searches"),
	}

	opts := workerutil.WorkerOptions{
		Name:              "exhaustive_search_worker",
		Description:       "runs the exhaustive search",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "exhaustive_search_worker"),
	}

	return dbworker.NewWorker[*exhaustive.Job](ctx, workerStore, handler, opts)
}

type exhaustiveSearch struct {
	logger log.Logger
}

var _ workerutil.Handler[*exhaustive.Job] = &exhaustiveSearch{}

func (e *exhaustiveSearch) Handle(ctx context.Context, logger log.Logger, record *exhaustive.Job) error {
	// TODO at the moment this does nothing. This will be implemented in a future PR.
	return errors.New("not implemented")
}

// InitExhaustiveSearchWorkerStore initializes and returns a dbworker.Store instance for the exhaustive search worker.
func InitExhaustiveSearchWorkerStore() (dbworkerstore.Store[*exhaustive.Job], error) {
	return initExhaustiveSearchWorkerStore.Init()
}

var initExhaustiveSearchWorkerStore = memo.NewMemoizedConstructor(func() (dbworkerstore.Store[*exhaustive.Job], error) {
	observationCtx := observation.NewContext(log.Scoped("store.exhaustive_search", "the exhaustive search worker store"))

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return exhaustive.NewJobWorkerStore(observationCtx, db.Handle()), nil
})
