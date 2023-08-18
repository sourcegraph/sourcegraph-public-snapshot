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

// NewExhaustiveSearchRepoWorker creates a background routine that periodically runs the exhaustive search of a repo.
func NewExhaustiveSearchRepoWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	workerStore dbworkerstore.Store[*types.ExhaustiveSearchRepoJob],
	exhaustiveSearchStore *store.Store,
) goroutine.BackgroundRoutine {
	handler := &exhaustiveSearchRepoHandler{
		logger: log.Scoped("exhaustive-search-repo", "The background worker running exhaustive searches on a repository"),
		store:  exhaustiveSearchStore,
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

type exhaustiveSearchRepoHandler struct {
	logger log.Logger
	store  *store.Store
}

var _ workerutil.Handler[*types.ExhaustiveSearchRepoJob] = &exhaustiveSearchRepoHandler{}

func (h *exhaustiveSearchRepoHandler) Handle(ctx context.Context, logger log.Logger, record *types.ExhaustiveSearchRepoJob) error {
	// TODO at the moment this does nothing. This will be implemented in a future PR.
	return errors.New("not implemented")
}
