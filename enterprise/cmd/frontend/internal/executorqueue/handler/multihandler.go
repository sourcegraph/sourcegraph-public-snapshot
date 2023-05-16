package handler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	executorstore "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/store"
	executortypes "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MultiHandler handles the HTTP requests of an executor for more than one queue. See ExecutorHandler for single-queue implementation.
type MultiHandler struct {
	JobTokenStore         executorstore.JobTokenStore
	CodeIntelQueueHandler QueueHandler[uploadsshared.Index]
	BatchesQueueHandler   QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]
	validQueues           []string
	RandomGenerator       RandomGenerator
	logger                log.Logger
}

// NewMultiHandler creates a new MultiHandler.
func NewMultiHandler(
	jobTokenStore executorstore.JobTokenStore,
	codeIntelQueueHandler QueueHandler[uploadsshared.Index],
	batchesQueueHandler QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob],
) MultiHandler {
	return MultiHandler{
		JobTokenStore:         jobTokenStore,
		CodeIntelQueueHandler: codeIntelQueueHandler,
		BatchesQueueHandler:   batchesQueueHandler,
		validQueues:           []string{codeIntelQueueHandler.Name, batchesQueueHandler.Name},
		RandomGenerator:       &realRandom{},
		logger:                log.Scoped("executor-multi-queue-handler", "The route handler for all executor queues"),
	}
}

func (m *MultiHandler) validateQueues(queues []string) []string {
	var invalidQueues []string
	for _, queue := range queues {
		if !slices.Contains(m.validQueues, queue) {
			invalidQueues = append(invalidQueues, queue)
		}
	}
	return invalidQueues
}

// ServeHTTP is the equivalent of ExecutorHandler.HandleDequeue for multiple queues.
func (m *MultiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload executortypes.DequeueRequest
	wrapHandler(w, r, &payload, m.logger, func() (int, any, error) {
		job, dequeued, err := m.dequeue(r.Context(), payload)
		if !dequeued {
			return http.StatusNoContent, nil, err
		}

		return http.StatusOK, job, err
	})
}

func (m *MultiHandler) dequeue(ctx context.Context, req executortypes.DequeueRequest) (executortypes.Job, bool, error) {
	if err := validateWorkerHostname(req.ExecutorName); err != nil {
		m.logger.Error(err.Error())
		return executortypes.Job{}, false, err
	}

	version2Supported := false
	if req.Version != "" {
		var err error
		version2Supported, err = api.CheckSourcegraphVersion(req.Version, "4.3.0-0", "2022-11-24")
		if err != nil {
			return executortypes.Job{}, false, err
		}
	}

	if invalidQueues := m.validateQueues(req.Queues); len(invalidQueues) > 0 {
		message := fmt.Sprintf("Invalid queue name(s) '%s' found. Supported queue names are '%s'.", strings.Join(invalidQueues, ", "), strings.Join(m.validQueues, ", "))
		m.logger.Error(message)
		return executortypes.Job{}, false, errors.New(message)
	}

	resourceMetadata := ResourceMetadata{
		NumCPUs:   req.NumCPUs,
		Memory:    req.Memory,
		DiskSpace: req.DiskSpace,
	}

	// Initialize the random number generator
	m.RandomGenerator.Seed(time.Now().UnixNano())

	// Shuffle the slice using the Fisher-Yates algorithm
	for i := len(req.Queues) - 1; i > 0; i-- {
		j := m.RandomGenerator.Intn(i + 1)
		req.Queues[i], req.Queues[j] = req.Queues[j], req.Queues[i]
	}

	logger := m.logger.Scoped("dequeue", "Pick a job record from the database.")
	var job executortypes.Job
	for _, queue := range req.Queues {
		switch queue {
		case m.BatchesQueueHandler.Name:
			record, dequeued, err := m.BatchesQueueHandler.Store.Dequeue(ctx, req.ExecutorName, nil)
			if err != nil {
				err = errors.Wrapf(err, "dbworkerstore.Dequeue %s", queue)
				logger.Error("Failed to dequeue", log.String("queue", queue), log.Error(err))
				return executortypes.Job{}, false, err
			}
			if !dequeued {
				// no batches job to dequeue, try next queue
				continue
			}

			job, err = m.BatchesQueueHandler.RecordTransformer(ctx, req.Version, record, resourceMetadata)
			if err != nil {
				markErr := markRecordAsFailed(ctx, m.BatchesQueueHandler.Store, record.RecordID(), err, logger)
				err = errors.Wrapf(errors.Append(err, markErr), "RecordTransformer %s", queue)
				logger.Error("Failed to transform record", log.String("queue", queue), log.Error(err))
				return executortypes.Job{}, false, err
			}
		case m.CodeIntelQueueHandler.Name:
			record, dequeued, err := m.CodeIntelQueueHandler.Store.Dequeue(ctx, req.ExecutorName, nil)
			if err != nil {
				err = errors.Wrapf(err, "dbworkerstore.Dequeue %s", queue)
				logger.Error("Failed to dequeue", log.String("queue", queue), log.Error(err))
				return executortypes.Job{}, false, err
			}
			if !dequeued {
				// no codeintel job to dequeue, try next queue
				continue
			}

			job, err = m.CodeIntelQueueHandler.RecordTransformer(ctx, req.Version, record, resourceMetadata)
			if err != nil {
				markErr := markRecordAsFailed(ctx, m.CodeIntelQueueHandler.Store, record.RecordID(), err, logger)
				err = errors.Wrapf(errors.Append(err, markErr), "RecordTransformer %s", queue)
				logger.Error("Failed to transform record", log.String("queue", queue), log.Error(err))
				return executortypes.Job{}, false, err
			}
		}
		if job.ID != 0 {
			job.Queue = queue
			break
		}
	}

	if job.ID == 0 {
		// all queues are empty, return nothing
		return executortypes.Job{}, false, nil
	}

	// If this executor supports v2, return a v2 payload. Based on this field,
	// marshalling will be switched between old and new payload.
	if version2Supported {
		job.Version = 2
	}

	logger = m.logger.Scoped("token", "Create or regenerate a job token.")
	token, err := m.JobTokenStore.Create(ctx, job.ID, job.Queue, job.RepositoryName)
	if err != nil {
		if errors.Is(err, executorstore.ErrJobTokenAlreadyCreated) {
			// Token has already been created, regen it.
			token, err = m.JobTokenStore.Regenerate(ctx, job.ID, job.Queue)
			if err != nil {
				err = errors.Wrap(err, "RegenerateToken")
				logger.Error("Failed to regenerate token", log.Error(err))
				return executortypes.Job{}, false, err
			}
		} else {
			err = errors.Wrap(err, "CreateToken")
			logger.Error("Failed to create token", log.Error(err))
			return executortypes.Job{}, false, err
		}
	}
	job.Token = token

	return job, true, nil
}

func markRecordAsFailed[T workerutil.Record](context context.Context, store dbworkerstore.Store[T], recordID int, err error, logger log.Logger) error {
	_, markErr := store.MarkFailed(context, recordID, fmt.Sprintf("failed to transform record: %s", err), dbworkerstore.MarkFinalOptions{})
	if markErr != nil {
		logger.Error("Failed to mark record as failed",
			log.Int("recordID", recordID),
			log.Error(markErr))
	}
	return markErr
}

// RandomGenerator is a wrapper for generating random numbers to support simple queue fairness.
// Its functions can be mocked out for consistent dequeuing in unit tests.
type RandomGenerator interface {
	Seed(seed int64)
	Intn(n int) int
}

type realRandom struct{}

func (r *realRandom) Seed(seed int64) {
	rand.Seed(seed)
}

func (r *realRandom) Intn(n int) int {
	return rand.Intn(n)
}
