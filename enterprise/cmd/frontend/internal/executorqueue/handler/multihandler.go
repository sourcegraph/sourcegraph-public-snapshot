package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	executorstore "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MultiHandler struct {
	JobTokenStore         executorstore.JobTokenStore
	CodeIntelQueueHandler QueueHandler[uploadsshared.Index]
	BatchesQueueHandler   QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]
	logger                log.Logger
}

func NewMultiHandler(
	jobTokenStore executorstore.JobTokenStore,
	codeIntelQueueHandler QueueHandler[uploadsshared.Index],
	batchesQueueHandler QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob],
) MultiHandler {
	return MultiHandler{
		JobTokenStore:         jobTokenStore,
		CodeIntelQueueHandler: codeIntelQueueHandler,
		BatchesQueueHandler:   batchesQueueHandler,
		logger:                log.Scoped("executor-multi-queue-handler", "The route handler for all executor queues"),
	}
}

var validQueues = []string{"batches", "codeintel"}

func validateQueues(queues []string) []string {
	var invalidQueues []string
	for _, queue := range queues {
		if !slices.Contains(validQueues, queue) {
			invalidQueues = append(invalidQueues, queue)
		}
	}
	return invalidQueues
}

// TODO: fairly sure this is basically executortypes.DequeueRequest with Queues extended
// (and WorkerHostName == ExecutorName?)
type dequeueRequest struct {
	Queues         []string `json:"queues"`
	WorkerHostName string   `json:"workerHostName"`
	Version        string   `json:"version"`
	NumCPUs        int      `json:"numCPUs,omitempty"`
	Memory         string   `json:"memory,omitempty"`
	DiskSpace      string   `json:"diskSpace,omitempty"`
}

func (m *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req dequeueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// TODO: should we also log errors here? Not sure
		http.Error(w, fmt.Sprintf("Failed to unmarshal payload: %s", err.Error()), http.StatusBadRequest)
	}

	// TODO: simply exported this method: I guess all of this will move into the handler package anyway so temp solution
	if err := validateWorkerHostname(req.WorkerHostName); err != nil {
		// TODO
	}

	version2Supported := false
	if req.Version != "" {
		var err error
		version2Supported, err = api.CheckSourcegraphVersion(req.Version, "4.3.0-0", "2022-11-24")
		if err != nil {
			// TODO: should we also log errors here? Not sure
			http.Error(w, fmt.Sprintf("Failed to check Sourcegraph version: %s", err.Error()), http.StatusInternalServerError)
		}
	}

	if invalidQueues := validateQueues(req.Queues); len(invalidQueues) != 0 {
		// TODO: should we also log errors here? Not sure
		http.Error(w, fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'. ", strings.Join(invalidQueues, ", "), strings.Join(validQueues, ", ")), http.StatusBadRequest)
	}

	resourceMetadata := ResourceMetadata{
		NumCPUs:   req.NumCPUs,
		Memory:    req.Memory,
		DiskSpace: req.DiskSpace,
	}
	var job executortypes.Job
	logger := log.Scoped("dequeue", "Select a job record from the database.")
	// TODO - impl fairness later
	for _, queue := range req.Queues {
		// TODO: basically replicating error handling of handler.dequeue() here
		switch queue {
		case "batches":
			record, _, err := m.BatchesQueueHandler.Store.Dequeue(r.Context(), req.WorkerHostName, nil)
			if err != nil {
				logger.Error("Handler returned an error", log.Error(err))
				http.Error(w, fmt.Sprintf("Failed to dequeue from queue %s: %s", queue, errors.Wrap(err, "dbworkerstore.Dequeue").Error()), http.StatusInternalServerError)
			}

			job, err = m.BatchesQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
			if err != nil {
				if _, err = m.BatchesQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); err != nil {
					logger.Error("Failed to mark record as failed",
						log.Int("recordID", record.RecordID()),
						log.Error(err))
				}

				http.Error(w, fmt.Sprintf("Failed to transform %s record into job: %s", queue, errors.Wrap(err, "RecordTransformer")), http.StatusInternalServerError)
			}
		case "codeintel":
			record, _, err := m.CodeIntelQueueHandler.Store.Dequeue(r.Context(), req.WorkerHostName, nil)
			if err != nil {
				logger.Error("Handler returned an error", log.Error(err))
				http.Error(w, fmt.Sprintf("Failed to dequeue from queue %s: %s", queue, errors.Wrap(err, "dbworkerstore.Dequeue").Error()), http.StatusInternalServerError)
			}
			job, err = m.CodeIntelQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
			if err != nil {
				if _, err = m.CodeIntelQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); err != nil {
					logger.Error("Failed to mark record as failed",
						log.Int("recordID", record.RecordID()),
						log.Error(err))
				}

				http.Error(w, fmt.Sprintf("Failed to transform %s record into job: %s", queue, errors.Wrap(err, "RecordTransformer")), http.StatusInternalServerError)
			}
		}
		if job.ID != 0 {
			break
		}
		// If this executor supports v2, return a v2 payload. Based on this field,
		// marshalling will be switched between old and new payload.
		if version2Supported {
			job.Version = 2
		}

		token, err := m.JobTokenStore.Create(r.Context(), job.ID, queue, job.RepositoryName)
		if err != nil {
			if errors.Is(err, executorstore.ErrJobTokenAlreadyCreated) {
				// Token has already been created, regen it.
				token, err = m.JobTokenStore.Regenerate(r.Context(), job.ID, queue)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to regenerate token: %s", errors.Wrap(err, "RegenerateToken").Error()), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, fmt.Sprintf("Failed to create token: %s", errors.Wrap(err, "CreateToken").Error()), http.StatusInternalServerError)
			}
		}
		job.Token = token
		job.Queue = queue
	}

	// TODO - does this actually work?
	if err := json.NewEncoder(w).Encode(job); err != nil {
		logger.Error("Failed to serialize payload", log.Error(err))
		http.Error(w, fmt.Sprintf("Failed to serialize payload: %s", err), http.StatusInternalServerError)
	}
}
