package common

import (
	"container/list"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	metricLabels = []string{"queue"}

	queueLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_gitserver_generic_queue_length",
		Help: "The number of items currently in the queue.",
	}, metricLabels)

	queueEnqueuedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_generic_queue_enqueued_total",
		Help: "The total number of items enqueued.",
	}, metricLabels)

	queueWaitTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_gitserver_generic_queue_wait_time_seconds",
		Help: "Time spent in queue waiting to be processed",
	}, metricLabels)

	queueProcessingTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_gitserver_generic_queue_processing_time_seconds",
	}, metricLabels)
)

var registerMetricsOnce = new(sync.Once)

func registerMetrics(observationCtx *observation.Context) {
	registerMetricsOnce.Do(func() {
		observationCtx.Registerer.MustRegister(
			queueLength,
			queueEnqueuedTotal,
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
	*metrics

	jobs *list.List

	mu sync.Mutex

	// FIXME: Make these private.
	// Coming soon in a follow up PR.
	Mutex sync.Mutex
	Cond  *sync.Cond
}

// NewQueue initializes a new Queue.
func NewQueue[T Jobber](obctx *observation.Context, name string, jobs *list.List) *Queue[T] {
	q := Queue[T]{jobs: jobs}
	q.Cond = sync.NewCond(&q.Mutex)

	// Register the metrics the first time this queue is used.
	registerMetrics(obctx)

	// Setup the metrics for this specific instance of the queue.
	q.metrics = &metrics{
		length:         queueLength.WithLabelValues(name),
		enqueuedTotal:  queueEnqueuedTotal.WithLabelValues(name),
		waitTime:       queueWaitTime.WithLabelValues(name),
		processingTime: queueProcessingTime.WithLabelValues(name),
	}

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
	q.length.Inc()
	q.enqueuedTotal.Inc()
}

// Pop returns the next job and a function that consumers of this job may use to record some
// metrics. If there's no next job available, it returns nil, nil.
func (q *Queue[T]) Pop() (*T, func(float64)) {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.jobs.Front()
	if next == nil {
		return nil, nil
	}

	job := q.jobs.Remove(next).(T)

	q.waitTime.Observe(time.Since(job.GetPushedAt()).Seconds())
	q.length.Dec()

	// NOTE: The function being returned is hardcoded at the moment. In the future this may be a
	// property of the queue if implementations need it. For now this is all we need.
	return &job, func(val float64) {
		q.processingTime.Observe(val)
	}
}

func (q *Queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.jobs.Len() == 0
}

type metrics struct {
	length         prometheus.Gauge
	enqueuedTotal  prometheus.Counter
	waitTime       prometheus.Observer
	processingTime prometheus.Observer
}
