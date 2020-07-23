package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// OperationMetrics contains three common metrics for any operation.
type OperationMetrics struct {
	Duration *prometheus.HistogramVec // How long did it take?
	Count    *prometheus.CounterVec   // How many things were processed?
	Errors   *prometheus.CounterVec   // How many errors occurred?
}

// Observe registers an observation of a single operation.
func (m *OperationMetrics) Observe(secs, count float64, err *error, lvals ...string) {
	if m == nil {
		return
	}

	m.Duration.WithLabelValues(lvals...).Observe(secs)
	m.Count.WithLabelValues(lvals...).Add(count)
	if err != nil && *err != nil {
		m.Errors.WithLabelValues(lvals...).Add(1)
	}
}

type operationMetricOptions struct {
	subsystem    string
	durationHelp string
	countHelp    string
	errorsHelp   string
	labels       []string
}

// OperationMetricsOption alter the default behavior of NewOperationMetrics.
type OperationMetricsOption func(o *operationMetricOptions)

// WithSubsystem overrides the default subsystem for all metrics.
func WithSubsystem(subsystem string) OperationMetricsOption {
	return func(o *operationMetricOptions) { o.subsystem = subsystem }
}

// WithDurationHelp overrides the default help text for duration metrics.
func WithDurationHelp(text string) OperationMetricsOption {
	return func(o *operationMetricOptions) { o.durationHelp = text }
}

// WithCountHelp overrides the default help text for count metrics.
func WithCountHelp(text string) OperationMetricsOption {
	return func(o *operationMetricOptions) { o.countHelp = text }
}

// WithErrorsHelp overrides the default help text for errors metrics.
func WithErrorsHelp(text string) OperationMetricsOption {
	return func(o *operationMetricOptions) { o.errorsHelp = text }
}

// WithLabels overrides the default labels for all metrics.
func WithLabels(labels ...string) OperationMetricsOption {
	return func(o *operationMetricOptions) { o.labels = labels }
}

// NewOperationMetrics creates an OperationMetrics value. The metrics will be
// immediately registered to the given registerer. This method panics on registration
// error. The supplied metricPrefix should be underscore_cased as it is used in the
// metric name.
func NewOperationMetrics(r prometheus.Registerer, metricPrefix string, fns ...OperationMetricsOption) *OperationMetrics {
	options := &operationMetricOptions{
		subsystem:    "",
		durationHelp: fmt.Sprintf("Time in seconds spent performing %s operations", metricPrefix),
		countHelp:    fmt.Sprintf("Total number of %s operations", metricPrefix),
		errorsHelp:   fmt.Sprintf("Total number of errors when performing %s operations", metricPrefix),
		labels:       nil,
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
		},
		options.labels,
	)
	r.MustRegister(duration)

	count := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "src",
			Name:      fmt.Sprintf("%s_total", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.countHelp,
		},
		options.labels,
	)
	r.MustRegister(count)

	errors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "src",
			Name:      fmt.Sprintf("%s_errors_total", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.errorsHelp,
		},
		options.labels,
	)
	r.MustRegister(errors)

	return &OperationMetrics{
		Duration: duration,
		Count:    count,
		Errors:   errors,
	}
}

type SingletonOperationMetrics struct {
	sync.Once
	metrics *OperationMetrics
}

// SingletonOperationMetrics returns an operation metrics instance. If no instance has been
// created yet, one is constructed with the given create function. This method is safe to
// access concurrently.
func (m *SingletonOperationMetrics) Get(create func() *OperationMetrics) *OperationMetrics {
	m.Do(func() {
		m.metrics = create()
	})
	return m.metrics
}
