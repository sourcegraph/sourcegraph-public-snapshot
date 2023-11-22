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

var registerMetricsOnce sync.Once

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

type queueItem[T any] struct {
	job      T
	pushedAt time.Time
}

// Queue is a threadsafe FIFO queue.
type Queue[T any] struct {
	*metrics

	jobs *list.List

	mu sync.Mutex

	// FIXME: Make these private.
	// Coming soon in a follow up PR.
	Mutex sync.Mutex
	Cond  *sync.Cond
}

// NewQueue initializes a new Queue.
func NewQueue[T any](obctx *observation.Context, name string, jobs *list.List) *Queue[T] {
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

	q.jobs.PushBack(&queueItem[T]{
		job:      job,
		pushedAt: time.Now(),
	})
	q.Cond.Signal()

	// Set the push time on the job's metadata. This will be used to observe the total wait time in
	// queue when this job is eventually popped.
	q.length.Inc()
	q.enqueuedTotal.Inc()
}

// Pop returns the next job and a function that consumers of this job may use to record some
// metrics. If there's no next job available, it returns nil, nil.
func (q *Queue[T]) Pop() (*T, func() time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()

	next := q.jobs.Front()
	if next == nil {
		return nil, nil
	}

	item := q.jobs.Remove(next).(*queueItem[T])

	q.waitTime.Observe(time.Since(item.pushedAt).Seconds())
	q.length.Dec()

	processingTime := time.Now()

	// NOTE: The function being returned is hardcoded at the moment. In the future this may be a
	// property of the queue if implementations need it. For now this is all we need.
	return &item.job, func() time.Duration {
		duration := time.Since(processingTime)
		q.processingTime.Observe(duration.Seconds())
		return duration
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
