package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewRepoMetadataWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.RepoMetadataWithName],
	gitClient gitserver.Client,
	observationContext *observation.Context,
) *workerutil.Worker[*btypes.RepoMetadataWithName] {
	options := workerutil.WorkerOptions{
		Name:              "batches_repo_metadata_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationContext, "batch_changes_repo_metadata"),
	}

	handler := &repoMetadataHandler{
		store:  s,
		client: gitClient,
	}

	worker := dbworker.NewWorker[*btypes.RepoMetadataWithName](ctx, workerStore, handler, options)
	return worker
}

type repoMetadataHandler struct {
	store  *store.Store
	client gitserver.Client
}

func (w *repoMetadataHandler) Handle(ctx context.Context, logger log.Logger, meta *btypes.RepoMetadataWithName) error {
	updated, err := service.CalculateRepoMetadata(ctx, w.client, service.CalculateRepoMetadataOpts{
		ID:   meta.RepoID,
		Name: meta.RepoName,
	})
	if err != nil {
		return errors.Wrap(err, "calculating repo metadata")
	}

	if err := w.store.UpsertRepoMetadata(ctx, updated); err != nil {
		return errors.Wrap(err, "upserting repo metadata")
	}

	return nil
}
