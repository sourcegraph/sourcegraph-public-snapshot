package server

import (
	"context"

	"github.com/hashicorp/go-multierror"
)

// heartbeat will release the transaction for any job that is not confirmed to be in-progress
// by the given executor. This method is called when the executor POSTs its in-progress job
// identifiers to the /heartbeat route. This method returns the set of identifiers which the
// executor erroneously claims to hold (and are sent back as a hint to stop processing).
func (h *handler) heartbeat(ctx context.Context, executorName string, jobIDs []int) ([]int, error) {
	deadJobs, unknownIDs, err := h.heartbeatJobs(ctx, executorName, jobIDs)
	err2 := h.requeueJobs(ctx, deadJobs)
	if err != nil && err2 != nil {
		err = multierror.Append(err, err2)
	}
	if err2 != nil {
		err = err2
	}
	return unknownIDs, err
}

// cleanup will release the transactions held by any executor that has not sent a heartbeat
// in a while. This method is called periodically in the background.
func (h *handler) cleanup(ctx context.Context) error {
	return h.requeueJobs(ctx, h.pruneExecutors())
}

// heartbeatJobs updates the set of job identifiers assigned to the given executor and returns
// any job that was known to us but not reported by the executor, plus the set of job identifiers
// reported by the executor which do not have an associated record held by this instance of the
// executor queue. This can occur when the executor-queue restarts and loses its in-memory state.
// We send these identifiers back to the executor as a hint to stop processing.
func (h *handler) heartbeatJobs(ctx context.Context, executorName string, ids []int) (dead []jobMeta, unknownIDs []int, errs error) {
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
		executorQueueIDsMap[job.record.RecordID()] = struct{}{}
		if _, ok := executorIDsMap[job.record.RecordID()]; ok || now.Sub(job.started) < h.options.UnreportedMaxAge {
			live = append(live, job)
			if err := h.heartbeatJob(ctx, job); err != nil {
				errs = multierror.Append(errs, err)
			}
		} else {
			dead = append(dead, job)
		}
	}
	executor.jobs = live
	executor.lastUpdate = now

	unknownIDs = make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := executorQueueIDsMap[id]; !ok {
			unknownIDs = append(unknownIDs, id)
		}
	}

	return dead, unknownIDs, errs
}

// pruneExecutors will release the transactions held by any executor that has not sent a
// heartbeat in a while and return the attached jobs.
func (h *handler) pruneExecutors() (jobs []jobMeta) {
	h.m.Lock()
	defer h.m.Unlock()

	for name, executor := range h.executors {
		if h.clock.Now().Sub(executor.lastUpdate) <= h.options.DeathThreshold {
			continue
		}

		jobs = append(jobs, executor.jobs...)
		delete(h.executors, name)
	}

	return jobs
}

func (h *handler) heartbeatJob(ctx context.Context, job jobMeta) error {
	queueOptions, ok := h.options.QueueOptions[job.queueName]
	if !ok {
		return ErrUnknownQueue
	}

	return queueOptions.Store.Heartbeat(ctx, job.record.RecordID())
}

// requeueJobs releases and requeues each of the given jobs.
func (h *handler) requeueJobs(ctx context.Context, jobs []jobMeta) (errs error) {
	for _, job := range jobs {
		if err := h.requeueJob(ctx, job); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

// requeueJob requeues the given job and releases the associated transaction.
func (h *handler) requeueJob(ctx context.Context, job jobMeta) error {
	queueOptions, ok := h.options.QueueOptions[job.queueName]
	if !ok {
		return ErrUnknownQueue
	}

	return queueOptions.Store.Requeue(ctx, job.record.RecordID(), h.clock.Now().Add(h.options.RequeueDelay))
}
