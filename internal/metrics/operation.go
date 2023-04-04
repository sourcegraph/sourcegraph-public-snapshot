package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// REDMetrics contains three common metrics for any operation.
// It is modeled after the RED method, which defines three characteristics for
// monitoring services:
//
//   - number (rate) of requests per second
//   - number of errors/failed operations
//   - amount of time per operation
//
// https://thenewstack.io/monitoring-microservices-red-method/.
type REDMetrics struct {
	Count    *prometheus.CounterVec   // How many things were processed?
	Errors   *prometheus.CounterVec   // How many errors occurred?
	Duration *prometheus.HistogramVec // How long did it take?
}

// Observe registers an observation of a single operation.
func (m *REDMetrics) Observe(secs, count float64, err *error, lvals ...string) {
	if m == nil {
		return
	}

	if err != nil && *err != nil {
		m.Errors.WithLabelValues(lvals...).Inc()
		m.Count.WithLabelValues(lvals...).Add(0)
	} else {
		m.Duration.WithLabelValues(lvals...).Observe(secs)
		m.Count.WithLabelValues(lvals...).Add(count)
	}
}

type redMetricOptions struct {
	subsystem       string
	durationHelp    string
	countHelp       string
	errorsHelp      string
	labels          []string
	durationBuckets []float64
}

// REDMetricsOption alter the default behavior of NewREDMetrics.
type REDMetricsOption func(o *redMetricOptions)

// WithSubsystem overrides the default subsystem for all metrics.
func WithSubsystem(subsystem string) REDMetricsOption {
	return func(o *redMetricOptions) { o.subsystem = subsystem }
}

// WithDurationHelp overrides the default help text for duration metrics.
func WithDurationHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.durationHelp = text }
}

// WithDurationBuckets overrides the default histogram bucket values for duration metrics.
func WithDurationBuckets(buckets []float64) REDMetricsOption {
	return func(o *redMetricOptions) {
		if len(buckets) != 0 {
			o.durationBuckets = buckets
		}
	}
}

// WithCountHelp overrides the default help text for count metrics.
func WithCountHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.countHelp = text }
}

// WithErrorsHelp overrides the default help text for errors metrics.
func WithErrorsHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.errorsHelp = text }
}

// WithLabels overrides the default labels for all metrics.
func WithLabels(labels ...string) REDMetricsOption {
	return func(o *redMetricOptions) { o.labels = labels }
}

// NewREDMetrics creates an REDMetrics value. The metrics will be
// immediately registered to the given registerer. This method panics on registration
// error. The supplied metricPrefix should be underscore_cased as it is used in the
// metric name.
func NewREDMetrics(r prometheus.Registerer, metricPrefix string, fns ...REDMetricsOption) *REDMetrics {
	options := &redMetricOptions{
		subsystem:       "",
		durationHelp:    fmt.Sprintf("Time in seconds spent performing successful %s operations", metricPrefix),
		countHelp:       fmt.Sprintf("Total number of successful %s operations", metricPrefix),
		errorsHelp:      fmt.Sprintf("Total number of %s operations resulting in an unexpected error", metricPrefix),
		labels:          nil,
		durationBuckets: prometheus.DefBuckets,
	}

	for _, fn := range fns {
		fn(options)
	}

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "src",
			Name:      fmt.Sprintf("%s_duration_seconds", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.durationHelp,
			Buckets:   options.durationBuckets,
		},
		options.labels,
	)
	duration = MustRegisterIgnoreDuplicate(r, duration)

	count := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "src",
			Name:      fmt.Sprintf("%s_total", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.countHelp,
		},
		options.labels,
	)
	count = MustRegisterIgnoreDuplicate(r, count)

	errors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "src",
			Name:      fmt.Sprintf("%s_errors_total", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.errorsHelp,
		},
		options.labels,
	)
	errors = MustRegisterIgnoreDuplicate(r, errors)

	return &REDMetrics{
		Duration: duration,
		Count:    count,
		Errors:   errors,
	}
}

// MustRegisterIgnoreDuplicate is like registerer.MustRegister(collector), except that it returns
// the already registered collector with the same ID if a duplicate collector is attempted to be
// registered.
func MustRegisterIgnoreDuplicate[T prometheus.Collector](registerer prometheus.Registerer, collector T) T {
	if err := registerer.Register(collector); err != nil {
		if e, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return e.ExistingCollector.(T)
		}
		panic(err) // otherwise, panic (as registerer.MustRegister would)
	}
	return collector
}

type SingletonREDMetrics struct {
	once    sync.Once
	metrics *REDMetrics
}

// Get returns a RED metrics instance. If no instance has been
// created yet, one is constructed with the given create function. This method is safe to
// access concurrently.
func (m *SingletonREDMetrics) Get(create func() *REDMetrics) *REDMetrics {
	m.once.Do(func() {
		m.metrics = create()
	})
	return m.metrics
}
