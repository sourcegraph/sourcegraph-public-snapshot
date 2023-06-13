package common

import (
	"container/list"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	queueLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_generic_queue_length",
		Help: "The number of items currently in the queue.",
	}, []string{"queue"})

	queueEnqueuedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_generic_queue_enqueued_total",
		Help: "The total number of items enqueued.",
	}, []string{"queue"})

	queueDequeuedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_generic_queue_dequeued_total",
		Help: "The total number of items dequeued.",
	}, []string{"queue"})

	queueWaitTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_gitserver_generic_queue_wait_time",
		Help: "Time spent in queue waiting to be processed",
	}, []string{"queue", "job_name"})

	queueProcessingTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_gitserver_generic_queue_processing_time",
	}, []string{"queue", "job_name"})
)

var registerMetricsOnce = new(sync.Once)

func registerMetrics(observationCtx *observation.Context) {
	registerMetricsOnce.Do(func() {
		observationCtx.Registerer.MustRegister(
			queueLength,
			queueEnqueuedTotal,
			queueDequeuedTotal,
			queueWaitTime,
			queueProcessingTime,
		)
	})
}

type Jobber interface {
	// Identifier returns a string that will help identify the job. It is **NOT** a unique ID for
	// every job, rather it must return the same string if the same job is pushed to the queue
	// twice.
	//
	// This is required in tracking metrics like wait time and processing time of a job.
	Identifier() string

	// GetPushedAt returns the pushedAt time from the JobMetadata.
	//
	// If it is invoked before SetPushedAt (see below), it will return the zero value of time.Time.
	GetPushedAt() time.Time

	// SetPushedAt sets the pushedAt time on the JobMetadata. It is idempotent.
	//
	// Invoking this the first time will set the current time and subsequent calls to this method
	// will be a NOOP.
	SetPushedAt()
}

// JobMetadata provides an API for managing metadata of a job. This is useful in metrics collection.
//
// Implementations of Jobber should embed this struct.
type JobMetadata struct {
	pushedAt     time.Time
	pushedAtOnce sync.Once
}

func (jm *JobMetadata) GetPushedAt() time.Time {
	return jm.pushedAt
}
func (jm *JobMetadata) SetPushedAt() {
	jm.pushedAtOnce.Do(func() {
		jm.pushedAt = time.Now()
	})
}

// Queue is a threadsafe FIFO queue.
type Queue[T Jobber] struct {
	// A name that uniquely identifies this queue. This is used in generating names for in-built
	// metrics supported by the queue.
	//
	// Hyphens will be replaced by underscores because hyphens are not allowed in the metric names.
	name string
	jobs *list.List

	mu sync.Mutex

	// FIXME: Make these private.
	// Coming soon in a follow up PR.
	Mutex sync.Mutex
	Cond  *sync.Cond
}

// NewQueue initializes a new Queue.
func NewQueue[T Jobber](obctx *observation.Context, name string, jobs *list.List) *Queue[T] {
	q := Queue[T]{
		name: name,
		jobs: jobs,
	}
	q.Cond = sync.NewCond(&q.Mutex)

	registerMetrics(obctx)

	return &q
}

// Push will queue the job to the end of the queue.
func (q *Queue[T]) Push(job T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs.PushBack(job)
	q.Cond.Signal()

	// Set the push time on the job's metadata. This will be used to observe the total wait time in
	// queue when this job is eventually popped.
	job.SetPushedAt()
	queueLength.WithLabelValues(q.name).Inc()
	queueEnqueuedTotal.WithLabelValues(q.name).Inc()
}

// Pop will return the next job. If there's no next job available, it returns nil.
func (q *Queue[T]) Pop() *T {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.jobs.Front()
	if next == nil {
		return nil
	}

	job := q.jobs.Remove(next).(T)

	queueWaitTime.WithLabelValues(q.name, job.Identifier()).Observe(
		time.Since(job.GetPushedAt()).Seconds(),
	)

	queueLength.WithLabelValues(q.name).Dec()
	queueDequeuedTotal.WithLabelValues(q.name).Inc()

	return &job
}

func (q *Queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.jobs.Len() == 0
}

func (q *Queue[T]) RecordProcessingTime(job T, start time.Time) {
	queueProcessingTime.WithLabelValues(q.name, job.Identifier()).Observe(time.Since(start).Seconds())
}
