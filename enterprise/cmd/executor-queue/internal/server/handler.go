package server

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	queueOptions QueueOptions
}

type QueueOptions struct {
	// Store is a required dbworker store store for each registered queue.
	Store store.Store

	// RecordTransformer is a required hook for each registered queue that transforms a generic
	// record from that queue into the job to be given to an executor.
	RecordTransformer func(ctx context.Context, record workerutil.Record) (apiclient.Job, error)
}

func newHandler(queueOptions QueueOptions) *handler {
	return &handler{
		queueOptions: queueOptions,
	}
}

var ErrUnknownJob = errors.New("unknown job")

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// a false-valued flag is returned.
func (h *handler) dequeue(ctx context.Context, executorName, executorHostname string) (_ apiclient.Job, dequeued bool, _ error) {
	// We explicitly DON'T want to use executorHostname here, it is NOT guaranteed to be unique.
	record, dequeued, err := h.queueOptions.Store.Dequeue(context.Background(), executorName, nil)
	if err != nil {
		return apiclient.Job{}, false, err
	}
	if !dequeued {
		return apiclient.Job{}, false, nil
	}

	job, err := h.queueOptions.RecordTransformer(ctx, record)
	if err != nil {
		if _, err := h.queueOptions.Store.MarkFailed(ctx, record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), store.MarkFinalOptions{}); err != nil {
			log15.Error("Failed to mark record as failed", "recordID", record.RecordID(), "error", err)
		}

		return apiclient.Job{}, false, err
	}

	return job, true, nil
}

// addExecutionLogEntry calls AddExecutionLogEntry for the given job.
func (h *handler) addExecutionLogEntry(ctx context.Context, executorName string, jobID int, entry workerutil.ExecutionLogEntry) error {
	return h.queueOptions.Store.AddExecutionLogEntry(ctx, jobID, entry, store.AddExecutionLogEntryOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting logs,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
		// We pass state to enforce adding log entries is only possible while the record is still dequeued.
		State: "processing",
	})
}

// markComplete calls MarkComplete for the given job.
func (h *handler) markComplete(ctx context.Context, executorName string, jobID int) error {
	ok, err := h.queueOptions.Store.MarkComplete(ctx, jobID, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if !ok {
		return ErrUnknownJob
	}
	return err
}

// markErrored calls MarkErrored for the given job.
func (h *handler) markErrored(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	ok, err := h.queueOptions.Store.MarkErrored(ctx, jobID, errorMessage, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if !ok {
		return ErrUnknownJob
	}
	return err
}

// markFailed calls MarkFailed for the given job.
func (h *handler) markFailed(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	ok, err := h.queueOptions.Store.MarkFailed(ctx, jobID, errorMessage, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if !ok {
		return ErrUnknownJob
	}
	return err
}
