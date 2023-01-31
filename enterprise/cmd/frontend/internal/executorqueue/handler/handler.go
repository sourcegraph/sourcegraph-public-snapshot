package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExecutorHandler handles the HTTP requests of an executor.
type ExecutorHandler interface {
	// Name is the name of the queue the handler processes.
	Name() string
	// AuthMiddleware is the specific auth middleware for the queue.
	AuthMiddleware(next http.Handler) http.Handler
	// HandleDequeue retrieves the next executor.Job to be processed in the queue.
	HandleDequeue(w http.ResponseWriter, r *http.Request)
	// HandleAddExecutionLogEntry adds the log entry for the executor.Job.
	HandleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	// HandleUpdateExecutionLogEntry updates the log entry for the executor.Job.
	HandleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request)
	// HandleMarkComplete updates the executor.Job to have a completed status.
	HandleMarkComplete(w http.ResponseWriter, r *http.Request)
	// HandleMarkErrored updates the executor.Job to have an errored status.
	HandleMarkErrored(w http.ResponseWriter, r *http.Request)
	// HandleMarkFailed updates the executor.Job to have a failed status.
	HandleMarkFailed(w http.ResponseWriter, r *http.Request)
	// HandleHeartbeat handles the heartbeat of an executor.
	HandleHeartbeat(w http.ResponseWriter, r *http.Request)
	// HandleCanceledJobs cancels the specified executor.Jobs.
	HandleCanceledJobs(w http.ResponseWriter, r *http.Request)
}

var _ ExecutorHandler = &handler[workerutil.Record]{}

type handler[T workerutil.Record] struct {
	queueHandler  QueueHandler[T]
	executorStore database.ExecutorStore
	jobTokenStore executor.JobTokenStore
	metricsStore  metricsstore.DistributedStore
	logger        log.Logger
}

// QueueHandler the specific logic for handling a queue.
type QueueHandler[T workerutil.Record] struct {
	// Name signifies the type of work the queue serves to executors.
	Name string
	// Store is a required dbworker store.
	Store store.Store[T]
	// RecordTransformer is a required hook for each registered queue that transforms a generic
	// record from that queue into the job to be given to an executor.
	RecordTransformer func(ctx context.Context, version string, record T, resourceMetadata ResourceMetadata) (executor.Job, error)
}

// NewHandler creates a new ExecutorHandler.
func NewHandler[T workerutil.Record](
	executorStore database.ExecutorStore,
	jobTokenStore executor.JobTokenStore,
	metricsStore metricsstore.DistributedStore,
	queueHandler QueueHandler[T],
) ExecutorHandler {
	return &handler[T]{
		executorStore: executorStore,
		jobTokenStore: jobTokenStore,
		metricsStore:  metricsStore,
		logger: log.Scoped(
			fmt.Sprintf("executor-queue-handler-%s", queueHandler.Name),
			fmt.Sprintf("The route handler for all executor %s dbworker API tunnel endpoints", queueHandler.Name),
		),
		queueHandler: queueHandler,
	}
}

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// a false-valued flag is returned.
func (h *handler[T]) dequeue(ctx context.Context, queueName string, metadata executorMetadata) (executor.Job, bool, error) {
	if err := validateWorkerHostname(metadata.name); err != nil {
		return executor.Job{}, false, err
	}

	version2Supported := false
	if metadata.version != "" {
		var err error
		version2Supported, err = api.CheckSourcegraphVersion(metadata.version, "4.3.0-0", "2022-11-24")
		if err != nil {
			return executor.Job{}, false, err
		}
	}

	// executorName is supposed to be unique.
	record, dequeued, err := h.queueHandler.Store.Dequeue(ctx, metadata.name, nil)
	if err != nil {
		return executor.Job{}, false, errors.Wrap(err, "dbworkerstore.Dequeue")
	}
	if !dequeued {
		return executor.Job{}, false, nil
	}

	logger := log.Scoped("dequeue", "Select a job record from the database.")
	job, err := h.queueHandler.RecordTransformer(ctx, metadata.version, record, metadata.resources)
	if err != nil {
		if _, err := h.queueHandler.Store.MarkFailed(ctx, record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), store.MarkFinalOptions{}); err != nil {
			logger.Error("Failed to mark record as failed",
				log.Int("recordID", record.RecordID()),
				log.Error(err))
		}

		return executor.Job{}, false, errors.Wrap(err, "RecordTransformer")
	}

	// If this executor supports v2, return a v2 payload. Based on this field,
	// marshalling will be switched between old and new payload.
	if version2Supported {
		job.Version = 2
	}

	token, err := h.jobTokenStore.Create(ctx, job.ID, queueName)
	if err != nil {
		// Maybe the token already exists (executor may have dies midway though processing the Job)?
		tokenExists, existsErr := h.jobTokenStore.Exists(ctx, job.ID, queueName)
		if existsErr != nil {
			return executor.Job{}, false, errors.CombineErrors(errors.Wrap(err, "CreateToken"), errors.Wrap(existsErr, "Exists"))
		}
		// The token does not exist AND we failed to created it. So bail out with the first error.
		if !tokenExists {
			return executor.Job{}, false, errors.Wrap(err, "CreateToken")
		}
		// The token does exist. Regenerate the token
		token, err = h.jobTokenStore.Regenerate(ctx, job.ID, queueName)
		if err != nil {
			return executor.Job{}, false, errors.Wrap(err, "Regenerate")
		}
	}
	job.Token = token

	return job, true, nil
}

type executorMetadata struct {
	name      string
	version   string
	resources ResourceMetadata
}

// ResourceMetadata is the specific resource data for an executor instance.
type ResourceMetadata struct {
	NumCPUs   int
	Memory    string
	DiskSpace string
}

// addExecutionLogEntry calls AddExecutionLogEntry for the given job.
func (h *handler[T]) addExecutionLogEntry(ctx context.Context, executorName string, jobID int, entry internalexecutor.ExecutionLogEntry) (int, error) {
	entryID, err := h.queueHandler.Store.AddExecutionLogEntry(ctx, jobID, entry, store.ExecutionLogEntryOptions{
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
func (h *handler[T]) updateExecutionLogEntry(ctx context.Context, executorName string, jobID int, entryID int, entry internalexecutor.ExecutionLogEntry) error {
	err := h.queueHandler.Store.UpdateExecutionLogEntry(ctx, jobID, entryID, entry, store.ExecutionLogEntryOptions{
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
func (h *handler[T]) markComplete(ctx context.Context, queueName string, executorName string, jobID int) error {
	ok, err := h.queueHandler.Store.MarkComplete(ctx, jobID, store.MarkFinalOptions{
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

	if err = h.jobTokenStore.Delete(ctx, jobID, queueName); err != nil {
		return errors.Wrap(err, "jobTokenStore.Delete")
	}

	return nil
}

// markErrored calls MarkErrored for the given job.
func (h *handler[T]) markErrored(ctx context.Context, queueName string, executorName string, jobID int, errorMessage string) error {
	ok, err := h.queueHandler.Store.MarkErrored(ctx, jobID, errorMessage, store.MarkFinalOptions{
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

	if err = h.jobTokenStore.Delete(ctx, jobID, queueName); err != nil {
		return errors.Wrap(err, "jobTokenStore.Delete")
	}

	return nil
}

// markFailed calls MarkFailed for the given job.
func (h *handler[T]) markFailed(ctx context.Context, queueName string, executorName string, jobID int, errorMessage string) error {
	ok, err := h.queueHandler.Store.MarkFailed(ctx, jobID, errorMessage, store.MarkFinalOptions{
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

	if err = h.jobTokenStore.Delete(ctx, jobID, queueName); err != nil {
		return errors.Wrap(err, "jobTokenStore.Delete")
	}

	return nil
}

// ErrUnknownJob
var ErrUnknownJob = errors.New("unknown job")

// heartbeat calls Heartbeat for the given jobs.
func (h *handler[T]) heartbeat(ctx context.Context, queueName string, executor types.Executor, ids []int) ([]int, []int, error) {
	if err := validateWorkerHostname(executor.Hostname); err != nil {
		return nil, nil, err
	}

	logger := log.Scoped("heartbeat", "Write this heartbeat to the database")

	// Write this heartbeat to the database so that we can populate the UI with recent executor activity.
	if err := h.executorStore.UpsertHeartbeat(ctx, executor); err != nil {
		logger.Error("Failed to upsert executor heartbeat", log.Error(err))
	}

	knownIDs, cancelIDs, err := h.queueHandler.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{
		// We pass the WorkerHostname, so the store enforces the record to be owned by this executor. When
		// the previous executor didn't report heartbeats anymore, but is still alive and reporting state,
		// both executors that ever got the job would be writing to the same record. This prevents it.
		WorkerHostname: executor.Hostname,
	})
	return knownIDs, cancelIDs, errors.Wrap(err, "dbworkerstore.UpsertHeartbeat")
}

// canceled reaches to the queueHandlers.FetchCanceled to determine jobs that need
// to be canceled.
// This endpoint is deprecated and should be removed in Sourcegraph 4.4.
func (h *handler[T]) canceled(ctx context.Context, queueName string, executorName string, knownIDs []int) ([]int, error) {
	if err := validateWorkerHostname(executorName); err != nil {
		return nil, err
	}
	// The Heartbeat method now handles both heartbeats and cancellation. For backcompat,
	// we fall back to this method.
	_, canceledIDs, err := h.queueHandler.Store.Heartbeat(ctx, knownIDs, store.HeartbeatOptions{
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
