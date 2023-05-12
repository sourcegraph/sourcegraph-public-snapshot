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
type dequeueRequest struct {
	Queues       []string `json:"queues"`
	ExecutorName string   `json:"executorName"`
	Version      string   `json:"version"`
	NumCPUs      int      `json:"numCPUs,omitempty"`
	Memory       string   `json:"memory,omitempty"`
	DiskSpace    string   `json:"diskSpace,omitempty"`
}

func (m *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req dequeueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("Failed to unmarshal payload"))
		m.logger.Error(err.Error())
		m.marshalAndRespondError(w, err)
		return
	}

	// TODO: simply exported this method: I guess all of this will move into the handler package anyway so temp solution
	if err := validateWorkerHostname(req.ExecutorName); err != nil {
		// TODO
	}

	version2Supported := false
	if req.Version != "" {
		var err error
		version2Supported, err = api.CheckSourcegraphVersion(req.Version, "4.3.0-0", "2022-11-24")
		if err != nil {
			m.marshalAndRespondError(w, err)
			return
		}
	}

	if invalidQueues := validateQueues(req.Queues); len(invalidQueues) != 0 {
		message := fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'. ", strings.Join(invalidQueues, ", "), strings.Join(validQueues, ", "))
		m.logger.Error(message)
		m.marshalAndRespondError(w, errors.New(message))
		return
	}

	resourceMetadata := ResourceMetadata{
		NumCPUs:   req.NumCPUs,
		Memory:    req.Memory,
		DiskSpace: req.DiskSpace,
	}

	logger := m.logger.Scoped("dequeue", "Select a job record from the database.")
	var job executortypes.Job
	// TODO - impl fairness later
	for _, queue := range req.Queues {
		// TODO: basically replicating error handling of handler.dequeue() here
		switch queue {
		case "batches":
			record, dequeued, err := m.BatchesQueueHandler.Store.Dequeue(r.Context(), req.ExecutorName, nil)
			if err != nil {
				err = errors.Wrapf(err, "dbworkerstore.Dequeue %s", queue)
				logger.Error("Handler returned an error", log.Error(err))
				m.marshalAndRespondError(w, err)
				return
			}
			if !dequeued {
				// no batches job to dequeue, try next queue
				continue
			}

			job, err = m.BatchesQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
			if err != nil {
				var markErr error
				if _, markErr = m.BatchesQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); markErr != nil {
					logger.Error("Failed to mark record as failed",
						log.String("queue", queue),
						log.Int("recordID", record.RecordID()),
						log.Error(markErr))
				}
				err = errors.Wrapf(errors.Append(err, markErr), "RecordTransformer %s", queue)
				m.marshalAndRespondError(w, err)
				return
			}
		case "codeintel":
			record, dequeued, err := m.CodeIntelQueueHandler.Store.Dequeue(r.Context(), req.ExecutorName, nil)
			if err != nil {
				err = errors.Wrapf(err, "dbworkerstore.Dequeue %s", queue)
				logger.Error("Handler returned an error", log.Error(err))
				m.marshalAndRespondError(w, err)
				return
			}
			if !dequeued {
				// no codeintel job to dequeue, try next queue
				continue
			}

			job, err = m.CodeIntelQueueHandler.RecordTransformer(r.Context(), req.Version, record, resourceMetadata)
			if err != nil {
				var markErr error
				if _, markErr = m.CodeIntelQueueHandler.Store.MarkFailed(r.Context(), record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{}); markErr != nil {
					logger.Error("Failed to mark record as failed",
						log.Int("recordID", record.RecordID()),
						log.Error(markErr))
				}
				err = errors.Wrapf(errors.Append(err, markErr), "RecordTransformer %s", queue)
				m.marshalAndRespondError(w, err)
				return
			}
		}
		if job.ID != 0 {
			job.Queue = queue
			break
		}
	}

	if job.ID == 0 {
		// all queues are empty, return no content
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// If this executor supports v2, return a v2 payload. Based on this field,
	// marshalling will be switched between old and new payload.
	if version2Supported {
		job.Version = 2
	}

	logger = m.logger.Scoped("token", "Create or regenerate a job token.")
	token, err := m.JobTokenStore.Create(r.Context(), job.ID, job.Queue, job.RepositoryName)
	if err != nil {
		if errors.Is(err, executorstore.ErrJobTokenAlreadyCreated) {
			// Token has already been created, regen it.
			token, err = m.JobTokenStore.Regenerate(r.Context(), job.ID, job.Queue)
			if err != nil {
				err = errors.Wrap(err, "RegenerateToken")
				logger.Error(err.Error())
				m.marshalAndRespondError(w, err)
				return
			}
		} else {
			err = errors.Wrap(err, "CreateToken")
			logger.Error(err.Error())
			m.marshalAndRespondError(w, err)
			return
		}
	}
	job.Token = token

	// TODO - does this actually work?
	if err := json.NewEncoder(w).Encode(job); err != nil {
		err = errors.Wrap(err, "Failed to serialize payload")
		m.logger.Error(err.Error())
		m.marshalAndRespondError(w, err)
	}
}

func (m *MultiHandler) marshalAndRespondError(w http.ResponseWriter, err error) {
	data, err := json.Marshal(errorResponse{Error: err.Error()})
	if err != nil {
		m.logger.Error("Failed to serialize payload", log.Error(err))
		data = []byte(fmt.Sprintf("Failed to serialize payload: %s", err))
	}
	http.Error(w, string(data), http.StatusInternalServerError)
}
