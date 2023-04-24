package multi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type multiHandler struct {
	multiQueueHandler MultiQueueHandler
	logger            log.Logger
}

type MultiQueueHandler struct {
	Handlers map[string]handler.ExecutorHandler
}

// TransformerFunc is the function to transform a workerutil.Record into an executor.Job.
type TransformerFunc[T workerutil.Record] func(ctx context.Context, version string, record T, resourceMetadata handler.ResourceMetadata) (executortypes.Job, error)

func NewMultiHandler(multiQueueHandler MultiQueueHandler) handler.ExecutorHandler {
	return &multiHandler{
		logger:            log.Scoped("executor-multi-queue-handler", "The generic route handler for all executor queues"),
		multiQueueHandler: multiQueueHandler,
	}
}

func (m *multiHandler) Name() string {
	//TODO implement me
	panic("implement me")
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

func (m *multiHandler) HandleDequeue(w http.ResponseWriter, r *http.Request) {
	var req dequeueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// TODO: should we also log errors here? Not sure
		http.Error(w, fmt.Sprintf("Failed to unmarshal payload: %s", err.Error()), http.StatusBadRequest)
	}

	if invalidQueues := validateQueues(req.Queues); len(invalidQueues) != 0 {
		// TODO: should we also log errors here? Not sure
		http.Error(w, fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'. ", strings.Join(invalidQueues, ", "), strings.Join(validQueues, ", ")), http.StatusBadRequest)
	}

	for _, queue := range req.Queues {
		// simulate as if the request was POSTed to /{queueName}/dequeue
		mux.Vars(r)["queueName"] = queue
		m.multiQueueHandler.Handlers[queue].HandleDequeue(w, r)
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

//func (m *multiHandler) dequeue(ctx context.Context, req dequeueRequest, w http.ResponseWriter) executortypes.Job {
//	// TODO: simply exported this method: I guess all of this will move into the handler package anyway so temp solution
//	if err := handler.ValidateWorkerHostname(req.WorkerHostName); err != nil {
//		// TODO
//	}
//
//	version2Supported := false
//	if req.Version != "" {
//		var err error
//		version2Supported, err = api.CheckSourcegraphVersion(req.Version, "4.3.0-0", "2022-11-24")
//		if err != nil {
//			// TODO: should we also log errors here? Not sure
//			http.Error(w, fmt.Sprintf("Failed to check Sourcegraph version: %s", err.Error()), http.StatusInternalServerError)
//		}
//	}
//
//	if invalidQueues := validateQueues(req.Queues); len(invalidQueues) != 0 {
//		// TODO: should we also log errors here? Not sure
//		http.Error(w, fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'. ", strings.Join(invalidQueues, ", "), strings.Join(validQueues, ", ")), http.StatusBadRequest)
//	}
//
//	resourceMetadata := handler.ResourceMetadata{
//		NumCPUs:   req.NumCPUs,
//		Memory:    req.Memory,
//		DiskSpace: req.DiskSpace,
//	}
//
//	logger := log.Scoped("dequeue", "Select a job record from the database.")
//	// TODO - impl fairness later
//	for _, queue := range req.Queues {
//		queueHandler := m.multiQueueHandler.QueueHandlers[queue]
//		// TODO: basically replicating error handling of handler.dequeue() here
//		record, _, err := queueHandler.Store.Dequeue(ctx, req.WorkerHostName, nil)
//		if err != nil {
//			logger.Error("Handler returned an error", log.Error(err))
//			http.Error(w, fmt.Sprintf("Failed to dequeue from queue %s: %s", queue, errors.Wrap(err, "dbworkerstore.Dequeue").Error()), http.StatusInternalServerError)
//		}
//		job, err := queueHandler.RecordTransformer(ctx, req.Version, record, resourceMetadata)
//		if err != nil {
//			if _, err = queueHandler.Store.MarkFailed(ctx, record.RecordID(), fmt.Sprintf("failed to transform record: %s", err), store.MarkFinalOptions{}); err != nil {
//				logger.Error("Failed to mark record as failed",
//					log.Int("recordID", record.RecordID()),
//					log.Error(err))
//			}
//
//			http.Error(w, fmt.Sprintf("Failed to transform %s record into job: %s", queue, errors.Wrap(err, "RecordTransformer")), http.StatusInternalServerError)
//		}
//		if job.ID != 0 {
//			break
//		}
//		// If this executor supports v2, return a v2 payload. Based on this field,
//		// marshalling will be switched between old and new payload.
//		if version2Supported {
//			job.Version = 2
//		}
//
//		token, err := m.jobTokenStore.Create(ctx, job.ID, queue, job.RepositoryName)
//		if err != nil {
//			if errors.Is(err, executorstore.ErrJobTokenAlreadyCreated) {
//				// Token has already been created, regen it.
//				token, err = m.jobTokenStore.Regenerate(ctx, job.ID, queue)
//				if err != nil {
//					http.Error(w, fmt.Sprintf("Failed to regenerate token: %s", errors.Wrap(err, "RegenerateToken").Error()), http.StatusInternalServerError)
//				}
//			} else {
//				http.Error(w, fmt.Sprintf("Failed to create token: %s", errors.Wrap(err, "CreateToken").Error()), http.StatusInternalServerError)
//			}
//		}
//		job.Token = token
//		job.Queue = queue
//
//		return job
//	}
//
//	//// TODO - does this actually work?
//	//if err := json.NewEncoder(w).Encode(job); err != nil {
//	//	logger.Error("Failed to serialize payload", log.Error(err))
//	//	http.Error(w, fmt.Sprintf("Failed to serialize payload: %s", err), http.StatusInternalServerError)
//	//}
//	return executortypes.Job{}
//}

func (m *multiHandler) HandleAddExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleUpdateExecutionLogEntry(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleMarkComplete(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleMarkErrored(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleMarkFailed(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (m *multiHandler) HandleCanceledJobs(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}
