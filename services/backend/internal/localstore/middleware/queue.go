package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
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
	} else {
		enqueued.WithLabelValues(j.Type).Inc()
	}
	return err
}

// LockJob implements store.Queue.
func (q *InstrumentedQueue) LockJob(ctx context.Context) (*store.LockedJob, error) {
	j, err := q.Queue.LockJob(ctx)
	if err != nil {
		errors.WithLabelValues("lockjob").Inc()
	} else if j != nil {
		lockedJobs.WithLabelValues(j.Type).Inc()
		return store.NewLockedJob(
			j.Job,
			func() error {
				err := j.MarkSuccess()
				if err != nil {
					errors.WithLabelValues("marksucess").Inc()
				} else {
					markedSuccess.WithLabelValues(j.Type).Inc()
				}
				return err
			},
			func(reason string) error {
				err := j.MarkError(reason)
				if err != nil {
					errors.WithLabelValues("markerror").Inc()
				} else {
					markedError.WithLabelValues(j.Type).Inc()
				}
				return err
			},
		), nil
	}
	return j, err
}

const (
	namespace = "src"
	typeLabel = "type"
)

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
