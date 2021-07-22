package server

import (
	"context"

	"github.com/hashicorp/go-multierror"
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
	now := h.clock.Now()

	executorIDsMap := map[int]struct{}{}
	for _, id := range ids {
		executorIDsMap[id] = struct{}{}
	}

	h.m.Lock()
	defer h.m.Unlock()

	executor, ok := h.executors[executorName]
	if !ok {
		executor = &executorMeta{}
		h.executors[executorName] = executor
	}

	executorQueueIDsMap := map[int]struct{}{}
	var live []jobMeta
	for _, job := range executor.jobs {
		executorQueueIDsMap[job.recordID] = struct{}{}
		if _, ok := executorIDsMap[job.recordID]; ok || now.Sub(job.started) < h.options.UnreportedMaxAge {
			live = append(live, job)
			if err := h.queueOptions.Store.Heartbeat(ctx, job.recordID); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	executor.jobs = live

	unknownInQueueJobs = make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := executorQueueIDsMap[id]; !ok {
			unknownInQueueJobs = append(unknownInQueueJobs, id)
		}
	}

	return unknownInQueueJobs, errs
}
