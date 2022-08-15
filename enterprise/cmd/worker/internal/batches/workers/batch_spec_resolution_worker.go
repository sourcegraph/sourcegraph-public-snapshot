package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewBatchSpecResolutionWorker creates a dbworker.newWorker that fetches BatchSpecResolutionJobs
// specs and passes them to the batchSpecWorkspaceCreator.
func NewBatchSpecResolutionWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	observationContext *observation.Context,
) *workerutil.Worker {
	e := &batchSpecWorkspaceCreator{
		store:  s,
		logger: log.Scoped("batch-spec-workspace-creator", "The background worker running workspace resolutions for batch changes"),
	}

	options := workerutil.WorkerOptions{
		Name:              "batch_changes_batch_spec_resolution_worker",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationContext, "batch_changes_batch_spec_resolution_worker"),
	}

	worker := dbworker.NewWorker(ctx, workerStore, e.HandlerFunc(), options)
	return worker
}
