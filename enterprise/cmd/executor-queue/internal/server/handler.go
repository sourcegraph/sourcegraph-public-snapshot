package server

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	options      Options
	queueOptions QueueOptions
}

type Options struct {
	// Port is the port on which to listen for HTTP connections.
	Port int
}

type QueueOptions struct {
	// Store is a required dbworker store store for each registered queue.
	Store store.Store

	// RecordTransformer is a required hook for each registered queue that transforms a generic
	// record from that queue into the job to be given to an executor.
	RecordTransformer func(ctx context.Context, record workerutil.Record) (apiclient.Job, error)
}

type executorMeta struct {
	jobs []jobMeta
}

type jobMeta struct {
	recordID int
	started  time.Time
}

func newHandler(options Options, queueOptions QueueOptions) *handler {
	return &handler{
		options:      options,
		queueOptions: queueOptions,
	}
}

var ErrUnknownJob = errors.New("unknown job")

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// or the server has hit its maximum transactions, a false-valued flag is returned.
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

// addExecutionLogEntry calls AddExecutionLogEntry for the given job. If the job identifier
// is not known, a false-valued flag is returned.
func (h *handler) addExecutionLogEntry(ctx context.Context, executorName string, jobID int, entry workerutil.ExecutionLogEntry) error {
	return h.queueOptions.Store.AddExecutionLogEntry(ctx, jobID, entry, store.AddExecutionLogEntryOptions{WorkerHostname: executorName})
}

// markComplete calls MarkComplete for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markComplete(ctx context.Context, executorName string, jobID int) error {
	_, err := h.queueOptions.Store.MarkComplete(ctx, jobID, store.MarkFinalOptions{WorkerHostname: executorName})
	return err
}

// markErrored calls MarkErrored for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markErrored(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	_, err := h.queueOptions.Store.MarkErrored(ctx, jobID, errorMessage, store.MarkFinalOptions{WorkerHostname: executorName})
	return err
}

// markFailed calls MarkFailed for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markFailed(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	_, err := h.queueOptions.Store.MarkFailed(ctx, jobID, errorMessage, store.MarkFinalOptions{WorkerHostname: executorName})
	return err
}

func (h *handler) scrapeMetrics() (numJobs int, numExecutors int) {
	var (
		JobIDs        []int
		ExecutorNames = make(map[string]struct{})
	)

	// TODO: Reimplement metrics scraping from the database.
	// SELECT
	// 	COUNT(id)
	// FROM
	// 	{table_name}
	// GROUP BY worker_hostname
	// for executorName, meta := range h.executors {
	// 	for _, job := range meta.jobs {
	// 		JobIDs = append(JobIDs, job.recordID)
	// 		ExecutorNames[executorName] = struct{}{}
	// 	}
	// }
	// TODO: We don't record executors anymore as we don't hold them in memory,
	// so we cannot tell when whether an executor is not getting a job or it is dead
	// right now. We want to build out a separate store for executors so they're persisted
	// for a while, so we can build some admin UI for viewing their status.
	// Maybe it is fine until then to not have this number and report it again then.
	// Otherwise we need to hold it in the DB now.

	return len(JobIDs), len(ExecutorNames)
}
