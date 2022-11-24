package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExecutorHandler interface {
	Name() string
	handleDequeue(w http.ResponseWriter, r *http.Request)
	handleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	handleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	handleMarkComplete(w http.ResponseWriter, r *http.Request)
	handleMarkErrored(w http.ResponseWriter, r *http.Request)
	handleMarkFailed(w http.ResponseWriter, r *http.Request)
	handleHeartbeat(w http.ResponseWriter, r *http.Request)
	handleCanceledJobs(w http.ResponseWriter, r *http.Request)
}

var _ ExecutorHandler = &handler[workerutil.Record]{}

type handler[T workerutil.Record] struct {
	QueueOptions[T]
	executorStore database.ExecutorStore
	metricsStore  metricsstore.DistributedStore
	logger        log.Logger
}

type QueueOptions[T workerutil.Record] struct {
	// Name signifies the type of work the queue serves to executors.
	Name string

	// Store is a required dbworker store store for each registered queue.
	Store store.Store[T]

	// RecordTransformer is a required hook for each registered queue that transforms a generic
	// record from that queue into the job to be given to an executor.
	RecordTransformer func(ctx context.Context, version string, record T, resourceMetadata ResourceMetadata) (apiclient.Job, error)
}

func NewHandler[T workerutil.Record](executorStore database.ExecutorStore, metricsStore metricsstore.DistributedStore, queueOptions QueueOptions[T]) *handler[T] {
	return &handler[T]{
		executorStore: executorStore,
		metricsStore:  metricsStore,
		logger:        log.Scoped("executor-queue-handler", "The route handler for all executor dbworker API tunnel endpoints"),
		QueueOptions:  queueOptions,
	}
}

var ErrUnknownJob = errors.New("unknown job")

type ResourceMetadata struct {
	NumCPUs   int
	Memory    string
	DiskSpace string
}

type executorMetadata struct {
	Name      string
	Version   string
	Resources ResourceMetadata
}

func (h *handler[T]) Name() string { return h.QueueOptions.Name }

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// a false-valued flag is returned.
func (h *handler[T]) dequeue(ctx context.Context, metadata executorMetadata) (_ apiclient.Job, dequeued bool, _ error) {
	if err := validateWorkerHostname(metadata.Name); err != nil {
		return apiclient.Job{}, false, err
	}

	version2Supported := false
	if metadata.Version != "" {
		var err error
		version2Supported, err = api.CheckSourcegraphVersion(metadata.Version, "4.3.0-0", "2022-11-24")
		if err != nil {
			return apiclient.Job{}, false, err
		}
	}

	// executorName is supposed to be unique.
	record, dequeued, err := h.Store.Dequeue(ctx, metadata.Name, nil)
	if err != nil {
		return apiclient.Job{}, false, errors.Wrap(err, "dbworkerstore.Dequeue")
	}
	if !dequeued {
		return apiclient.Job{}, false, nil
	}

	logger := log.Scoped("dequeue", "Select a job record from the database.")
	job, err := h.RecordTransformer(ctx, metadata.Version, record, metadata.Resources)
	if err != nil {
		if _, err := h.Store.MarkFailed(ctx, record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), store.MarkFinalOptions{}); err != nil {
			logger.Error("Failed to mark record as failed",
				log.Int("recordID", record.RecordID()),
				log.Error(err))
		}

		return apiclient.Job{}, false, errors.Wrap(err, "RecordTransformer")
	}

	// If this executor supports v2, return a v2 payload. Based on this field,
	// marshalling will be switched between old and new payload.
	if version2Supported {
		job.Version = 2
	}

	return job, true, nil
}

// addExecutionLogEntry calls AddExecutionLogEntry for the given job.
func (h *handler[T]) addExecutionLogEntry(ctx context.Context, executorName string, jobID int, entry workerutil.ExecutionLogEntry) (entryID int, err error) {
	if err := validateWorkerHostname(executorName); err != nil {
		return 0, err
	}

	entryID, err = h.Store.AddExecutionLogEntry(ctx, jobID, entry, store.ExecutionLogEntryOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting logs,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
		// We pass state to enforce adding log entries is only possible while the record is still dequeued.
		State: "processing",
	})
	if err == store.ErrExecutionLogEntryNotUpdated {
		return 0, ErrUnknownJob
	}
	return entryID, errors.Wrap(err, "dbworkerstore.AddExecutionLogEntry")
}

// updateExecutionLogEntry calls UpdateExecutionLogEntry for the given job and entry.
func (h *handler[T]) updateExecutionLogEntry(ctx context.Context, executorName string, jobID int, entryID int, entry workerutil.ExecutionLogEntry) error {
	if err := validateWorkerHostname(executorName); err != nil {
		return err
	}

	err := h.Store.UpdateExecutionLogEntry(ctx, jobID, entryID, entry, store.ExecutionLogEntryOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting logs,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
		// We pass state to enforce adding log entries is only possible while the record is still dequeued.
		State: "processing",
	})
	if err == store.ErrExecutionLogEntryNotUpdated {
		return ErrUnknownJob
	}
	return errors.Wrap(err, "dbworkerstore.UpdateExecutionLogEntry")
}

// markComplete calls MarkComplete for the given job.
func (h *handler[T]) markComplete(ctx context.Context, executorName string, jobID int) error {
	if err := validateWorkerHostname(executorName); err != nil {
		return err
	}

	ok, err := h.Store.MarkComplete(ctx, jobID, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.MarkComplete")
	}
	if !ok {
		return ErrUnknownJob
	}
	return nil
}

// markErrored calls MarkErrored for the given job.
func (h *handler[T]) markErrored(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	if err := validateWorkerHostname(executorName); err != nil {
		return err
	}

	ok, err := h.Store.MarkErrored(ctx, jobID, errorMessage, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.MarkErrored")
	}
	if !ok {
		return ErrUnknownJob
	}
	return nil
}

// markFailed calls MarkFailed for the given job.
func (h *handler[T]) markFailed(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	if err := validateWorkerHostname(executorName); err != nil {
		return err
	}

	ok, err := h.Store.MarkFailed(ctx, jobID, errorMessage, store.MarkFinalOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.MarkFailed")
	}
	if !ok {
		return ErrUnknownJob
	}
	return nil
}

// heartbeat calls Heartbeat for the given jobs.
func (h *handler[T]) heartbeat(ctx context.Context, executor types.Executor, ids []int) (knownIDs, cancelIDs []int, err error) {
	if err := validateWorkerHostname(executor.Hostname); err != nil {
		return nil, nil, err
	}

	logger := log.Scoped("heartbeat", "Write this heartbeat to the database")

	// Write this heartbeat to the database so that we can populate the UI with recent executor activity.
	if err := h.executorStore.UpsertHeartbeat(ctx, executor); err != nil {
		logger.Error("Failed to upsert executor heartbeat", log.Error(err))
	}

	knownIDs, cancelIDs, err = h.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executor.Hostname,
	})
	return knownIDs, cancelIDs, errors.Wrap(err, "dbworkerstore.UpsertHeartbeat")
}

// canceled reaches to the queueOptions.FetchCanceled to determine jobs that need
// to be canceled.
// This endpoint is deprecated and should be removed in Sourcegraph 4.4.
func (h *handler[T]) canceled(ctx context.Context, executorName string, knownIDs []int) (canceledIDs []int, err error) {
	if err := validateWorkerHostname(executorName); err != nil {
		return nil, err
	}
	// The Heartbeat method now handles both heartbeats and cancellation. For backcompat,
	// we fall back to this method.
	_, canceledIDs, err = h.Store.Heartbeat(ctx, knownIDs, store.HeartbeatOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executorName,
	})
	return canceledIDs, errors.Wrap(err, "dbworkerstore.CanceledJobs")
}

// validateWorkerHostname validates the WorkerHostname field sent for all the endpoints.
// We don't allow empty hostnames, as it would bypass the hostname verification, which
// could lead to stray workers updating records they no longer own.
func validateWorkerHostname(hostname string) error {
	if hostname == "" {
		return errors.New("worker hostname cannot be empty")
	}
	return nil
}
