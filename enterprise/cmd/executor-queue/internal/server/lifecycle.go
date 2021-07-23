package server

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// TODO: Fix this comment.
// heartbeat will release the transaction for any job that is not confirmed to be in-progress
// by the given executor. This method is called when the executor POSTs its in-progress job
// identifiers to the /heartbeat route. This method returns the set of identifiers which the
// executor erroneously claims to hold (and are sent back as a hint to stop processing).
// heartbeatJobs updates the set of job identifiers assigned to the given executor and returns
// any job that was known to us but not reported by the executor, plus the set of job identifiers
// reported by the executor which do not have an associated record held by this instance of the
// executor queue. This can occur when the executor-queue restarts and loses its in-memory state.
// We send these identifiers back to the executor as a hint to stop processing.
func (h *handler) heartbeat(ctx context.Context, executorName string, ids []int) (unknownInQueueJobs []int, errs error) {
	knownIDs, err := h.queueOptions.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{WorkerHostname: executorName})
	if err != nil {
		return nil, err
	}
	knownIDsMap := map[int]struct{}{}
	for _, id := range knownIDs {
		knownIDsMap[id] = struct{}{}
	}

	unknownInQueueJobs = make([]int, 0)
	for _, id := range ids {
		if _, ok := knownIDsMap[id]; !ok {
			unknownInQueueJobs = append(unknownInQueueJobs, id)
		}
	}

	return unknownInQueueJobs, errs
}
