package server

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// heartbeat is called when the executor POSTs its in-progress job identifiers to the /heartbeat route.
// This method returns the set of identifiers which the executor erroneously claims to hold (and are sent
// back as a hint to stop processing).
// The set of job identifiers assigned to the given executor are getting a heartbeat, indicating they're
// still being worked on.
func (h *handler) heartbeat(ctx context.Context, executorName string, ids []int) (unknownInQueueJobs []int, errs error) {
	knownIDs, err := h.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{WorkerHostname: executorName})
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
