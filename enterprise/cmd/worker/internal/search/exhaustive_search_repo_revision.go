package search

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewExhaustiveSearchRepoRevisionWorker creates a background routine that periodically runs the exhaustive search of a revision on a repo.
func NewExhaustiveSearchRepoRevisionWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*types.ExhaustiveSearchRepoRevisionJob],
	exhaustiveSearchStore *store.Store,
) goroutine.BackgroundRoutine {
	handler := &exhaustiveSearchRepoRevHandler{
		logger: log.Scoped("exhaustive-search-repo-revision", "The background worker running exhaustive searches on a revision of a repository"),
		store:  exhaustiveSearchStore,
	}

	opts := workerutil.WorkerOptions{
		Name:              "exhaustive_search_repo_revision_worker",
		Description:       "runs the exhaustive search on a revision of a repository",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "exhaustive_search_repo_revision_worker"),
	}

	return dbworker.NewWorker[*types.ExhaustiveSearchRepoRevisionJob](ctx, workerStore, handler, opts)
}

type exhaustiveSearchRepoRevHandler struct {
	logger log.Logger
	store  *store.Store
}

var _ workerutil.Handler[*types.ExhaustiveSearchRepoRevisionJob] = &exhaustiveSearchRepoRevHandler{}

func (h *exhaustiveSearchRepoRevHandler) Handle(ctx context.Context, logger log.Logger, record *types.ExhaustiveSearchRepoRevisionJob) error {
	// TODO at the moment this does nothing. This will be implemented in a future PR.
	return errors.New("not implemented")
}
