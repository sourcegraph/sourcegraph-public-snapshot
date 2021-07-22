package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/derision-test/glock"
	"github.com/inconshreveable/log15"

	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	options      Options
	queueOptions QueueOptions
	queueName    string
	clock        glock.Clock
	executors    map[string]*executorMeta
	m            sync.Mutex // protects executors
	queueMetrics *QueueMetrics
}

type Options struct {
	// Port is the port on which to listen for HTTP connections.
	Port int

	// UnreportedMaxAge is the maximum time between a record being dequeued and it appearing
	// in the executor's heartbeat requests before it being considered lost.
	UnreportedMaxAge time.Duration
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
	queueName string
	recordID  int
	started   time.Time
}

func newHandlerWithMetrics(options Options, queueOptions QueueOptions, queueName string, clock glock.Clock, observationContext *observation.Context) *handler {
	return &handler{
		options:      options,
		queueOptions: queueOptions,
		queueName:    queueName,
		clock:        clock,
		executors:    map[string]*executorMeta{},
		queueMetrics: newQueueMetrics(observationContext),
	}
}

var (
	ErrUnknownQueue = errors.New("unknown queue")
	ErrUnknownJob   = errors.New("unknown job")
)

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// or the server has hit its maximum transactions, a false-valued flag is returned.
func (h *handler) dequeue(ctx context.Context, executorName, executorHostname string) (_ apiclient.Job, dequeued bool, _ error) {
	record, dequeued, err := h.queueOptions.Store.Dequeue(context.Background(), executorHostname, nil)
	if err != nil {
		return apiclient.Job{}, false, err
	}
	if !dequeued {
		return apiclient.Job{}, false, nil
	}

	job, err := h.queueOptions.RecordTransformer(ctx, record)
	if err != nil {
		if _, err := h.queueOptions.Store.MarkFailed(ctx, record.RecordID(), fmt.Sprintf("failed to transform record: %s", err)); err != nil {
			log15.Error("Failed to mark record as failed", "recordID", record.RecordID(), "error", err)
		}

		return apiclient.Job{}, false, err
	}

	now := h.clock.Now()
	h.addMeta(executorName, jobMeta{queueName: h.queueName, recordID: record.RecordID(), started: now})
	return job, true, nil
}

// addExecutionLogEntry calls AddExecutionLogEntry for the given job. If the job identifier
// is not known, a false-valued flag is returned.
func (h *handler) addExecutionLogEntry(ctx context.Context, executorName string, jobID int, entry workerutil.ExecutionLogEntry) error {
	_, err := h.findMeta(h.queueName, executorName, jobID, false)
	if err != nil {
		return err
	}

	if err := h.queueOptions.Store.AddExecutionLogEntry(ctx, jobID, entry); err != nil {
		return err
	}

	return nil
}

// markComplete calls MarkComplete for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markComplete(ctx context.Context, executorName string, jobID int) error {
	job, err := h.findMeta(h.queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	_, err = h.queueOptions.Store.MarkComplete(ctx, job.recordID)
	return err
}

// markErrored calls MarkErrored for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markErrored(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	job, err := h.findMeta(h.queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	_, err = h.queueOptions.Store.MarkErrored(ctx, job.recordID, errorMessage)
	return err
}

// markFailed calls MarkFailed for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (h *handler) markFailed(ctx context.Context, executorName string, jobID int, errorMessage string) error {
	job, err := h.findMeta(h.queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	_, err = h.queueOptions.Store.MarkFailed(ctx, job.recordID, errorMessage)
	return err
}

// findMeta returns the job with the given id and executor name. If the job is
// unknown, an error is returned. If the remove parameter is true, the job will
// be removed from the executor's job list on success.
func (h *handler) findMeta(queueName, executorName string, jobID int, remove bool) (jobMeta, error) {
	h.m.Lock()
	defer h.m.Unlock()

	executor, ok := h.executors[executorName]
	if !ok {
		return jobMeta{}, ErrUnknownJob
	}

	for i, job := range executor.jobs {
		if job.queueName == queueName && job.recordID == jobID {
			if remove {
				l := len(executor.jobs) - 1
				executor.jobs[i] = executor.jobs[l]
				executor.jobs = executor.jobs[:l]
				h.updateMetrics()
			}

			return job, nil
		}
	}

	return jobMeta{}, ErrUnknownJob
}

// addMeta adds a job to the given executor's job list.
func (h *handler) addMeta(executorName string, job jobMeta) {
	h.m.Lock()
	defer h.m.Unlock()

	executor, ok := h.executors[executorName]
	if !ok {
		executor = &executorMeta{}
		h.executors[executorName] = executor
	}

	executor.jobs = append(executor.jobs, job)
	h.updateMetrics()
}

func (h *handler) updateMetrics() {
	type queueStat struct {
		JobIDs        []int
		ExecutorNames map[string]struct{}
	}
	queueStats := map[string]queueStat{}

	for executorName, meta := range h.executors {
		for _, job := range meta.jobs {
			stat, ok := queueStats[job.queueName]
			if !ok {
				stat = queueStat{
					ExecutorNames: map[string]struct{}{},
				}
			}

			stat.JobIDs = append(stat.JobIDs, job.recordID)
			stat.ExecutorNames[executorName] = struct{}{}
			queueStats[job.queueName] = stat
		}
	}

	for queueName, temp := range queueStats {
		h.queueMetrics.NumJobs.WithLabelValues(queueName).Set(float64(len(temp.JobIDs)))
		h.queueMetrics.NumExecutors.WithLabelValues(queueName).Set(float64(len(temp.ExecutorNames)))
	}
}
