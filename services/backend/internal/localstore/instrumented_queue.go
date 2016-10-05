package localstore

import (
	"sync"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/inconshreveable/log15.v2"
)

// instrumentedQueue wraps queue to instrument logging and metrics.
type instrumentedQueue struct{}

// Enqueue implements Queue.
func (q *instrumentedQueue) Enqueue(ctx context.Context, j *Job) error {
	err := (&queue{}).Enqueue(ctx, j)
	if err != nil {
		queueErrors.WithLabelValues("enqueue").Inc()
		log15.Debug("queue.Enqueue failed", "type", j.Type, "err", err)
	} else {
		queueEnqueued.WithLabelValues(j.Type).Inc()
		log15.Debug("queue.Enqueue success", "type", j.Type)
	}
	return err
}

// LockJob implements Queue.
func (q *instrumentedQueue) LockJob(ctx context.Context) (*LockedJob, error) {
	j, err := (&queue{}).LockJob(ctx)
	if err != nil {
		queueErrors.WithLabelValues("lockjob").Inc()
		log15.Debug("queue.LockJob failed", "err", err)
	} else if j != nil {
		queueLockedJobs.WithLabelValues(j.Type).Inc()
		log15.Debug("queue.LockJob success", "type", j.Type)
		return NewLockedJob(
			j.Job,
			func() error {
				err := j.MarkSuccess()
				if err != nil {
					queueErrors.WithLabelValues("marksucess").Inc()
					log15.Debug("LockedJob.MarkSuccess failed", "type", j.Type, "err", err)
				} else {
					queueMarkedSuccess.WithLabelValues(j.Type).Inc()
					log15.Debug("LockedJob.MarkSuccess success", "type", j.Type)
				}
				return err
			},
			func(reason string) error {
				err := j.MarkError(reason)
				if err != nil {
					queueErrors.WithLabelValues("markerror").Inc()
					log15.Debug("LockedJob.MarkError failed", "type", j.Type, "reason", reason, "err", err)
				} else {
					queueMarkedError.WithLabelValues(j.Type).Inc()
					log15.Debug("LockedJob.MarkError success", "type", j.Type, "reason", reason)
				}
				return err
			},
		), nil
	}
	return j, err
}

// Stats implements Queue.
func (q *instrumentedQueue) Stats(ctx context.Context) (map[string]QueueStats, error) {
	return (&queue{}).Stats(ctx)
}

const (
	namespace = "src"
	typeLabel = "type"
)

func init() {
	prometheus.MustRegister(newQueueStatsCollector())
}

// newQueueStatsCollector returns a prometheus collector based on the
// statistics returned by queue.Stats(). ctx needs to be long lived, and will
// be used when calling queue.Stats()
func newQueueStatsCollector() prometheus.Collector {
	return &queueStatsCollector{
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
			Help:      "The number of jobs in the queue (including running jobs) which have previously been queueMarkedError.",
		}, []string{typeLabel}),
	}
}

type queueStatsCollector struct {
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
	stats, err := (&queue{}).Stats(context.Background())
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

var queueEnqueued = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "enqueue_total",
	Help:      "Total number of Jobs successfully enqueued.",
}, []string{typeLabel})
var queueLockedJobs = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "lockedjobs_total",
	Help:      "Total number of Jobs successfully Locked.",
}, []string{typeLabel})
var queueErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "errors_total",
	Help:      "Total number of errors.",
}, []string{"method"})
var queueMarkedSuccess = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "marked_success_total",
	Help:      "Total number of LockedJobs.MarkSuccess().",
}, []string{typeLabel})
var queueMarkedError = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: "queue",
	Name:      "marked_error_total",
	Help:      "Total number of LockedJobs.MarkError().",
}, []string{typeLabel})

func init() {
	prometheus.MustRegister(queueEnqueued)
	prometheus.MustRegister(queueLockedJobs)
	prometheus.MustRegister(queueErrors)
	prometheus.MustRegister(queueMarkedSuccess)
	prometheus.MustRegister(queueMarkedError)
}
