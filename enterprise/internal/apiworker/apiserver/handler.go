package apiserver

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/efritz/glock"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	options          Options
	clock            glock.Clock
	executors        map[string]*executorMeta
	dequeueSemaphore chan struct{} // tracks available dequeue slots
	m                sync.Mutex    // protects executors
}

type Options struct {
	// Port is the port on which to listen for HTTP connections.
	Port int

	// QueueOptions is a map from queue name to options specific to that queue.
	QueueOptions map[string]QueueOptions

	// MaximumNumTransactions is the maximum number of active records that can be given out
	// to executors from this machine. The dequeue method will stop returning records while
	// the number of outstanding transactions is at or above this threshold.
	MaximumNumTransactions int

	// RequeueDelay controls how far into the future to make a job record visible to the job
	// queue once the currently processing executor has become unresponsive.
	RequeueDelay time.Duration

	// UnreportedMaxAge is the maximum time between a record being dequeued and it appearing
	// in the executor's heartbeat requests before it being considered lost.
	UnreportedMaxAge time.Duration

	// DeathThreshold is the minimum time since the last heartbeat of an executor before that
	// executor can be considered as unresponsive. This should be configured to be longer than
	// the duration between heartbeat interval.
	DeathThreshold time.Duration

	// CleanupInterval is the duration between periodic invocations of Cleanup, which will
	// requeue any records that are "lost" according to the thresholds described above.
	CleanupInterval time.Duration
}

type QueueOptions struct {
	// Store is a required dbworker store store for each registered queue.
	Store store.Store

	// RecordTransformer is a required hook for each registered queue that transforms a generic
	// record from that queue into the job to be given to an executor.
	RecordTransformer func(record workerutil.Record) (apiclient.Job, error)
}

type executorMeta struct {
	lastUpdate time.Time
	jobs       []jobMeta
}

type jobMeta struct {
	queueName string
	record    workerutil.Record
	tx        store.Store
	started   time.Time
}

func newHandler(options Options, clock glock.Clock) *handler {
	dequeueSemaphore := make(chan struct{}, options.MaximumNumTransactions)
	for i := 0; i < options.MaximumNumTransactions; i++ {
		dequeueSemaphore <- struct{}{}
	}

	return &handler{
		options:          options,
		clock:            clock,
		dequeueSemaphore: dequeueSemaphore,
		executors:        map[string]*executorMeta{},
	}
}

var ErrUnknownQueue = errors.New("unknown queue")
var ErrUnknownJob = errors.New("unknown job")

// dequeue selects a job record from the database and stashes metadata including
// the job record and the locking transaction. If no job is available for processing,
// or the server has hit its maximum transactions, a false-valued flag is returned.
func (m *handler) dequeue(ctx context.Context, queueName, executorName string) (_ apiclient.Job, dequeued bool, _ error) {
	queueOptions, ok := m.options.QueueOptions[queueName]
	if !ok {
		return apiclient.Job{}, false, ErrUnknownQueue
	}

	select {
	case <-m.dequeueSemaphore:
	default:
		return apiclient.Job{}, false, nil
	}
	defer func() {
		if !dequeued {
			// Ensure that if we do not dequeue a record successfully we do not
			// leak from the semaphore. This will happen if the dequeue call fails
			// or if there are no records to process
			m.dequeueSemaphore <- struct{}{}
		}
	}()

	record, tx, dequeued, err := queueOptions.Store.DequeueWithIndependentTransactionContext(ctx, nil)
	if err != nil {
		return apiclient.Job{}, false, err
	}
	if !dequeued {
		return apiclient.Job{}, false, nil
	}

	job, err := queueOptions.RecordTransformer(record)
	if err != nil {
		return apiclient.Job{}, false, tx.Done(err)
	}

	now := m.clock.Now()
	m.addMeta(executorName, jobMeta{queueName: queueName, record: record, tx: tx, started: now})
	return job, true, nil
}

// addExecutionLogEntry calls AddExecutionLogEntry for the given job. If the job identifier
// is not known, a false-valued flag is returned.
func (m *handler) addExecutionLogEntry(ctx context.Context, queueName, executorName string, jobID int, entry workerutil.ExecutionLogEntry) error {
	job, err := m.findMeta(queueName, executorName, jobID, false)
	if err != nil {
		return err
	}

	if err := job.tx.AddExecutionLogEntry(ctx, jobID, entry); err != nil {
		return err
	}

	return nil
}

// markComplete calls MarkComplete for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (m *handler) markComplete(ctx context.Context, queueName, executorName string, jobID int) error {
	job, err := m.findMeta(queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	defer func() { m.dequeueSemaphore <- struct{}{} }()

	_, err = job.tx.MarkComplete(ctx, job.record.RecordID())

	if doneErr := job.tx.Done(err); doneErr != err {
		return doneErr
	}

	return nil
}

// markErrored calls MarkErrored for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (m *handler) markErrored(ctx context.Context, queueName, executorName string, jobID int, errorMessage string) error {
	job, err := m.findMeta(queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	defer func() { m.dequeueSemaphore <- struct{}{} }()

	_, err = job.tx.MarkErrored(ctx, job.record.RecordID(), errorMessage)

	if doneErr := job.tx.Done(err); doneErr != err {
		return doneErr
	}

	return nil
}

// markFailed calls MarkFailed for the given job, then commits the job's transaction.
// The job is removed from the executor's job list on success.
func (m *handler) markFailed(ctx context.Context, queueName, executorName string, jobID int, errorMessage string) error {
	job, err := m.findMeta(queueName, executorName, jobID, true)
	if err != nil {
		return err
	}

	defer func() { m.dequeueSemaphore <- struct{}{} }()

	_, err = job.tx.MarkFailed(ctx, job.record.RecordID(), errorMessage)

	if doneErr := job.tx.Done(err); doneErr != err {
		return doneErr
	}

	return nil
}

// findMeta returns the job with the given id and executor name. If the job is
// unknown, an error is returned. If the remove parameter is true, the job will
// be removed from the executor's job list on success.
func (m *handler) findMeta(queueName, executorName string, jobID int, remove bool) (jobMeta, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if _, ok := m.options.QueueOptions[queueName]; !ok {
		return jobMeta{}, ErrUnknownQueue
	}

	executor, ok := m.executors[executorName]
	if !ok {
		return jobMeta{}, ErrUnknownJob
	}

	for i, job := range executor.jobs {
		if job.queueName == queueName && job.record.RecordID() == jobID {
			if remove {
				l := len(executor.jobs) - 1
				executor.jobs[i] = executor.jobs[l]
				executor.jobs = executor.jobs[:l]
			}

			return job, nil
		}
	}

	return jobMeta{}, ErrUnknownJob
}

// addMeta adds a job to the given executor's job list.
func (m *handler) addMeta(executorName string, job jobMeta) {
	m.m.Lock()
	defer m.m.Unlock()

	executor, ok := m.executors[executorName]
	if !ok {
		executor = &executorMeta{}
		m.executors[executorName] = executor
	}

	now := m.clock.Now()
	executor.jobs = append(executor.jobs, job)
	executor.lastUpdate = now
}
