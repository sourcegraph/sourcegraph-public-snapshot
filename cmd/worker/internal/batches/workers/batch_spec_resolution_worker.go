package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewBatchSpecResolutionWorker creates a dbworker.newWorker that fetches BatchSpecResolutionJobs
// specs and passes them to the batchSpecWorkspaceCreator.
func NewBatchSpecResolutionWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.BatchSpecResolutionJob],
) *workerutil.Worker[*btypes.BatchSpecResolutionJob] {
	e := &batchSpecWorkspaceCreator{
		store:  s,
		logger: log.Scoped("batch-spec-workspace-creator"),
	}

	options := workerutil.WorkerOptions{
		Name:              "batch_changes_batch_spec_resolution_worker",
		Description:       "runs the workspace resolutions for batch specs, for batch changes running server-side",
		NumHandlers:       5,
		Interval:          1 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "batch_changes_batch_spec_resolution_worker"),
	}

	worker := dbworker.NewWorker[*btypes.BatchSpecResolutionJob](ctx, workerStore, e.HandlerFunc(), options)
	return worker
}
