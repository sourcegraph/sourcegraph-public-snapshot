package middleware

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

// InstrumentedQueue wraps q to instrument logging and metrics.
type InstrumentedQueue struct {
	store.Queue
}

var _ store.Queue = (*InstrumentedQueue)(nil)

// Enqueue implements store.Queue.
func (q *InstrumentedQueue) Enqueue(ctx context.Context, j *store.Job) error {
	err := q.Queue.Enqueue(ctx, j)
	if err != nil {
		errors.WithLabelValues("enqueue").Inc()
		log15.Debug("queue.Enqueue failed", "type", j.Type, "err", err)
	} else {
		enqueued.WithLabelValues(j.Type).Inc()
		log15.Debug("queue.Enqueue success", "type", j.Type)
	}
	return err
}

// LockJob implements store.Queue.
func (q *InstrumentedQueue) LockJob(ctx context.Context) (*store.LockedJob, error) {
	j, err := q.Queue.LockJob(ctx)
	if err != nil {
		errors.WithLabelValues("lockjob").Inc()
		log15.Debug("queue.LockJob failed", "err", err)
	} else if j != nil {
		lockedJobs.WithLabelValues(j.Type).Inc()
		log15.Debug("queue.LockJob success", "type", j.Type)
		return store.NewLockedJob(
			j.Job,
			func() error {
				err := j.MarkSuccess()
				if err != nil {
					errors.WithLabelValues("marksucess").Inc()
					log15.Debug("LockedJob.MarkSuccess failed", "type", j.Type, "err", err)
				} else {
					markedSuccess.WithLabelValues(j.Type).Inc()
					log15.Debug("LockedJob.MarkSuccess success", "type", j.Type)
				}
				return err
			},
			func(reason string) error {
				err := j.MarkError(reason)
				if err != nil {
					errors.WithLabelValues("markerror").Inc()
					log15.Debug("LockedJob.MarkError failed", "type", j.Type, "reason", reason, "err", err)
				} else {
					markedError.WithLabelValues(j.Type).Inc()
					log15.Debug("LockedJob.MarkError success", "type", j.Type, "reason", reason)
				}
				return err
			},
		), nil
	}
	return j, err
}

// Stats implements store.Queue.
func (q *InstrumentedQueue) Stats(ctx context.Context) (map[string]store.QueueStats, error) {
	return q.Queue.Stats(ctx)
}

const (
	namespace = "src"
	typeLabel = "type"
)

// NewQueueStatsCollector returns a prometheus collector based on the
// statistics returned by queue.Stats(). ctx needs to be long lived, and will
// be used when calling queue.Stats()
func NewQueueStatsCollector(ctx context.Context, queue store.Queue) prometheus.Collector {
	return &queueStatsCollector{
		queue: queue,
		ctx:   ctx,
		types: make(map[string]bool),
		numJobs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "queue",
			Name:      "jobs",
			Help:      "The number of jobs in the queue (including running jobs).",
		}, []string{typeLabel}),
		numJobsWithError: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "queue",
			Name:      "jobs_with_error",
			Help:      "The number of jobs in the queue (including running jobs) which have previously been MarkedError.",
		}, []string{typeLabel}),
	}
}

type queueStatsCollector struct {
	queue store.Queue
	ctx   context.Context

	types            map[string]bool
	mu               sync.Mutex
	numJobs          *prometheus.GaugeVec
	numJobsWithError *prometheus.GaugeVec
}

// Describe implements prometheus.Collector
func (c *queueStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.numJobs.Describe(ch)
	c.numJobsWithError.Describe(ch)
}

// Collect implements prometheus.Collector
func (c *queueStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := c.queue.Stats(c.ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		log15.Debug("Error collecting queue stats", "err", err)
	} else {
		for t := range stats {
			c.types[t] = true
		}
		for t := range c.types {
			s := stats[t]
			c.numJobs.WithLabelValues(t).Set(float64(s.NumJobs))
			c.numJobsWithError.WithLabelValues(t).Set(float64(s.NumJobsWithError))
		}
	}
	c.numJobs.Collect(ch)
	c.numJobsWithError.Collect(ch)
	c.numJobs.Reset()
	c.numJobsWithError.Reset()
}

var enqueued = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "enqueue_total",
	Help:      "Total number of Jobs successfully enqueued.",
}, []string{typeLabel})
var lockedJobs = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "lockedjobs_total",
	Help:      "Total number of Jobs successfully Locked.",
}, []string{typeLabel})
var errors = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "errors_total",
	Help:      "Total number of errors.",
}, []string{"method"})
var markedSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "marked_success_total",
	Help:      "Total number of LockedJobs.MarkSuccess().",
}, []string{typeLabel})
var markedError = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "marked_error_total",
	Help:      "Total number of LockedJobs.MarkError().",
}, []string{typeLabel})

func init() {
	prometheus.MustRegister(enqueued)
	prometheus.MustRegister(lockedJobs)
	prometheus.MustRegister(errors)
	prometheus.MustRegister(markedSuccess)
	prometheus.MustRegister(markedError)
}
