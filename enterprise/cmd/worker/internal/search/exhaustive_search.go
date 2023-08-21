package search

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewExhaustiveSearchWorker creates a background routine that periodically runs the exhaustive search.
func NewExhaustiveSearchWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*types.ExhaustiveSearchJob],
	exhaustiveSearchStore *store.Store,
	searcher service.NewSearcher,
) goroutine.BackgroundRoutine {
	handler := &ExhaustiveSearchHandler{
		Store:    exhaustiveSearchStore,
		Searcher: searcher,
	}

	opts := workerutil.WorkerOptions{
		Name:              "exhaustive_search_worker",
		Description:       "runs the exhaustive search",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "exhaustive_search_worker"),
	}

	return dbworker.NewWorker[*types.ExhaustiveSearchJob](ctx, workerStore, handler, opts)
}

// ExhaustiveSearchHandler is a handler that runs the exhaustive search.
type ExhaustiveSearchHandler struct {
	// Store is the store used by the handler to interact with the database.
	Store *store.Store
	// Searcher is the searcher used by the handler to run the search.
	Searcher service.NewSearcher
}

var _ workerutil.Handler[*types.ExhaustiveSearchJob] = &ExhaustiveSearchHandler{}

func (h *ExhaustiveSearchHandler) Handle(ctx context.Context, logger log.Logger, record *types.ExhaustiveSearchJob) error {
	logger.Debug(
		"Handling exhaustive search job",
		log.Int64("id", record.ID),
		log.String("query", record.Query),
	)

	search, err := h.Searcher.NewSearch(ctx, record.Query)
	if err != nil {
		return errors.Wrap(err, "failed to create new search")
	}

	specs, err := search.RepositoryRevSpecs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get repository revision specifications")
	}

	// TODO: should this be done in a transaction?
	for _, spec := range specs {
		logger.Debug(
			"Handling exhaustive search job",
			log.Int64("id", record.ID),
			log.Int32("repository", int32(spec.Repository)),
			log.String("revisionSpecifier", spec.RevisionSpecifier),
		)
		job := types.ExhaustiveSearchRepoJob{
			SearchJobID: record.ID,
			RepoID:      spec.Repository,
			RefSpec:     spec.RevisionSpecifier,
		}
		if _, err = h.Store.CreateExhaustiveSearchRepoJob(ctx, job); err != nil {
			// TODO: is this really a failure? Should we just log it?
			return errors.Wrapf(err, "failed to create exhaustive search repo job for repository %d", spec.Repository)
		}
	}

	return nil
}
