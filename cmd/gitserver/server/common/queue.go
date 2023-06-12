package common

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Jobber interface {
	// Identifier returns a string that will help identify the job. It is **NOT** a unique ID for
	// every job, rather it must return the same string if the same job is pushed to the queue
	// twice.
	//
	// This is required in tracking metrics like wait time and processing time of a job.
	Identifier() string

	// UUID returns a unique identifier for the job. If the same job is pushed twice to the queue,
	// this will return a unique string for each of those jobs.
	UUID() string
}

// Queue is a threadsafe FIFO queue.
type Queue[T Jobber] struct {
	// metrics provides in-built support for observability of the queue.
	*metrics

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
//
// IMPORTANT: name must be unique or else setting up metrics will panic
// due to a duplicate name.
func NewQueue[T Jobber](ctx *observation.Context, name string, jobs *list.List) *Queue[T] {
	q := Queue[T]{
		// In case a consumer uses hyphens in the name, replace them with underscores because this
		// name is used to generate metric names and we want to generate consistent name in our metrics.
		name: strings.Replace(name, "-", "_", -1),
		jobs: jobs,
	}
	q.Cond = sync.NewCond(&q.Mutex)
	q.registerMetrics(ctx)

	return &q
}

// Push will queue the job to the end of the queue.
func (q *Queue[T]) Push(job T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.jobs.PushBack(job)
	q.Cond.Signal()

	q.length.Inc()
	q.waitTime.WithLabelValues(job.UUID(), job.Identifier(), "push").Observe(float64(time.Now().Unix()))
	q.enqueuedTotal.Inc()
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

	q.length.Dec()
	q.waitTime.WithLabelValues(job.UUID(), job.Identifier(), "pop").Observe(float64(time.Now().Unix()))
	q.dequeuedTotal.Inc()

	return &job
}

func (q *Queue[T]) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.jobs.Len() == 0
}

func (q *Queue[T]) RecordProcessingTime(jobIdentifier string, start time.Time) {
	q.processingTime.WithLabelValues(jobIdentifier).Observe(
		time.Since(start).Seconds(),
	)
}

func (q *Queue[T]) metricName(name string) string {
	return fmt.Sprintf("src_%s_%s", q.name, name)
}

type metrics struct {
	length        prometheus.Gauge
	enqueuedTotal prometheus.Counter
	dequeuedTotal prometheus.Counter

	waitTime       *prometheus.HistogramVec
	processingTime *prometheus.HistogramVec
}

func (q *Queue[T]) registerMetrics(observationCtx *observation.Context) {
	length := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: q.metricName("length"),
		Help: "The number of items currently in the queue.",
	})
	observationCtx.Registerer.MustRegister(length)

	enqueuedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: q.metricName("enqueued_total"),
		Help: "The total number of items enqueued.",
	})
	observationCtx.Registerer.MustRegister(enqueuedTotal)

	dequeuedTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: q.metricName("dequeued_total"),
		Help: "The total number of items dequeued.",
	})
	observationCtx.Registerer.MustRegister(dequeuedTotal)

	waitTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: q.metricName("wait_time"),
		Help: "Time spent in queue waiting to be processed",
	}, []string{"job_uuid", "job_identifier", "event_type"})
	observationCtx.Registerer.MustRegister(waitTime)

	processingTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: q.metricName("processing_time"),
	}, []string{"job_identifier"})
	observationCtx.Registerer.MustRegister(processingTime)

	q.metrics = &metrics{
		length:         length,
		enqueuedTotal:  enqueuedTotal,
		dequeuedTotal:  dequeuedTotal,
		waitTime:       waitTime,
		processingTime: processingTime,
	}
}
