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

// NewExhaustiveSearchRepoWorker creates a background routine that periodically runs the exhaustive search of a repo.
func NewExhaustiveSearchRepoWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*types.ExhaustiveSearchRepoJob],
	exhaustiveSearchStore *store.Store,
	searcher service.NewSearcher,
) goroutine.BackgroundRoutine {
	handler := &ExhaustiveSearchRepoHandler{
		Store:    exhaustiveSearchStore,
		Searcher: searcher,
	}

	opts := workerutil.WorkerOptions{
		Name:              "exhaustive_search_repo_worker",
		Description:       "runs the exhaustive search on a repository",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "exhaustive_search_repo_worker"),
	}

	return dbworker.NewWorker[*types.ExhaustiveSearchRepoJob](ctx, workerStore, handler, opts)
}

// ExhaustiveSearchRepoHandler is a handler that runs the exhaustive search on a repository.
type ExhaustiveSearchRepoHandler struct {
	// Store is the store used by the handler to interact with the database.
	Store *store.Store
	// Searcher is the searcher used by the handler to run the search.
	Searcher service.NewSearcher
}

var _ workerutil.Handler[*types.ExhaustiveSearchRepoJob] = &ExhaustiveSearchRepoHandler{}

func (h *ExhaustiveSearchRepoHandler) Handle(ctx context.Context, logger log.Logger, record *types.ExhaustiveSearchRepoJob) error {
	logger.Debug(
		"Handling exhaustive search repo job",
		log.Int64("id", record.ID),
		log.Int32("repo", int32(record.RepoID)),
		log.String("refSpec", record.RefSpec),
	)

	originalSearchJob, err := h.Store.GetExhaustiveSearchJobByID(ctx, record.SearchJobID)
	if err != nil {
		return errors.Wrap(err, "getting exhaustive search job")
	}

	search, err := h.Searcher.NewSearch(ctx, originalSearchJob.Query)
	if err != nil {
		return errors.Wrap(err, "creating search")
	}

	revSpec := service.RepositoryRevSpec{
		Repository:        record.RepoID,
		RevisionSpecifier: record.RefSpec,
	}
	specs, err := search.ResolveRepositoryRevSpec(ctx, revSpec)
	if err != nil {
		return errors.Wrap(err, "resolving repository rev spec")
	}

	// TODO: should this be done in a transaction?
	for _, spec := range specs {
		logger.Debug(
			"Creating exhaustive search repo revision job",
			log.Int64("repoJobID", record.ID),
			log.Int32("repository", int32(spec.Repository)),
			log.String("revision", spec.Revision),
		)

		job := types.ExhaustiveSearchRepoRevisionJob{
			SearchRepoJobID: record.ID,
			Revision:        spec.Revision,
		}
		if _, err = h.Store.CreateExhaustiveSearchRepoRevisionJob(ctx, job); err != nil {
			// TODO: should this be an error or just log it?
			return errors.Wrapf(err, "creating exhaustive search repo revision job for revision %q", spec.Revision)
		}
	}

	return nil
}
