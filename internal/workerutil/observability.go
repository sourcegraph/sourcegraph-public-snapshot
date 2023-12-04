package workerutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type WorkerObservability struct {
	// logger is the root logger provided for observability. Prefer to use a more granular
	// logger provided by operations where relevant.
	logger log.Logger

	// temporary solution to have configurable trace ahead-of-time sample for worker jobs
	// to avoid swamping sinks with traces.
	traceSampler func(job Record) bool

	operations *operations
	numJobs    Gauge
}

type Gauge interface {
	Inc()
	Dec()
}

type operations struct {
	handle     *observation.Operation
	postHandle *observation.Operation
	preHandle  *observation.Operation
}

type observabilityOptions struct {
	labels          map[string]string
	durationBuckets []float64
	// temporary solution to have configurable trace ahead-of-time sample for worker jobs
	// to avoid swamping sinks with traces.
	traceSampler func(job Record) bool
}

type ObservabilityOption func(o *observabilityOptions)

func WithSampler(fn func(job Record) bool) func(*observabilityOptions) {
	return func(o *observabilityOptions) { o.traceSampler = fn }
}

func WithLabels(labels map[string]string) ObservabilityOption {
	return func(o *observabilityOptions) { o.labels = labels }
}

func WithDurationBuckets(buckets []float64) ObservabilityOption {
	return func(o *observabilityOptions) { o.durationBuckets = buckets }
}

// NewMetrics creates and registers the following metrics for a generic worker instance.
//
//   - {prefix}_duration_seconds_bucket: handler operation latency histogram
//   - {prefix}_total: number of handler operations
//   - {prefix}_error_total: number of handler operations resulting in an error
//   - {prefix}_handlers: the number of active handler routines
//
// The given labels are emitted on each metric. If WithSampler option is not passed,
// traces will have a 1 in 2 probability of being sampled.
func NewMetrics(observationCtx *observation.Context, prefix string, opts ...ObservabilityOption) WorkerObservability {
	options := &observabilityOptions{
		durationBuckets: prometheus.DefBuckets,
		traceSampler: func(job Record) bool {
			return rand.Int31()%2 == 0
		},
	}

	for _, fn := range opts {
		fn(options)
	}

	keys := make([]string, 0, len(options.labels))
	values := make([]string, 0, len(options.labels))
	for key, value := range options.labels {
		keys = append(keys, key)
		values = append(values, value)
	}

	gauge := func(name, help string) prometheus.Gauge {
		gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("src_%s_%s", prefix, name),
			Help: help,
		}, keys)

		// TODO(sqs): TODO(single-binary): Ideally we would be using MustRegister here, not the
		// IgnoreDuplicate variant. This is a bit of a hack to allow 2 executor instances to run in a
		// single binary deployment.
		gaugeVec = metrics.MustRegisterIgnoreDuplicate(observationCtx.Registerer, gaugeVec)
		return gaugeVec.WithLabelValues(values...)
	}

	numJobs := gauge(
		"handlers",
		"The number of active handlers.",
	)

	return WorkerObservability{
		logger:       observationCtx.Logger,
		traceSampler: options.traceSampler,
		operations:   newOperations(observationCtx, prefix, keys, values, options.durationBuckets),
		numJobs:      newLenientConcurrencyGauge(numJobs, time.Second*5),
	}
}

func newOperations(observationCtx *observation.Context, prefix string, keys, values []string, durationBuckets []float64) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		prefix,
		metrics.WithLabels(append(keys, "op")...),
		metrics.WithCountHelp("Total number of method invocations."),
		metrics.WithDurationBuckets(durationBuckets),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              name,
			MetricLabelValues: append(append([]string{}, values...), name),
			Metrics:           redMetrics,
		})
	}

	return &operations{
		handle:     op("Handle"),
		postHandle: op("PostHandle"),
		preHandle:  op("PreHandle"),
	}
}

// newLenientConcurrencyGauge creates a new gauge-like object that
// emits the maximum value over the last five seconds into the given
// gauge. Note that this gauge should be used to track concurrency
// only, meaning that running the gauge into the negatives may produce
// unwanted behavior.
//
// This method begins an immortal background routine.
//
// This gauge should be used to smooth-over the randomness sampled by
// Prometheus by emitting the aggregate we'll likely be using with this
// type of data directly.
//
// Without wrapping concurrency gauges in this object, we tend to sample
// zero values consistently when the underlying resource is only occupied
// for a small amount of time (e.g., less than 500ms). We attribute this
// to random Prometheus samplying alignments.
func newLenientConcurrencyGauge(gauge prometheus.Gauge, interval time.Duration) Gauge {
	ch := make(chan float64)
	go runLenientConcurrencyGauge(gauge, ch, interval)

	return &channelGauge{ch: ch}
}

func runLenientConcurrencyGauge(gauge prometheus.Gauge, ch <-chan float64, interval time.Duration) {
	value := float64(0)                // The current value
	max := float64(0)                  // The max value in the current window
	ticker := time.NewTicker(interval) // The window over which to track the max value
	reset := true                      // Whether the next read of ch should reset the max

	for {
		select {
		case <-ticker.C:
			gauge.Set(max)
			reset = true

		case update, ok := <-ch:
			if !ok {
				return
			}

			if reset {
				// We've already emitted the max for the previous window, but we don't
				// reset max to zero immediately after updating the gauge. That tends
				// to emit zero values if our ticker frequency is less than our channel
				// read frequency.

				max = 0
				reset = false
			}

			value += update
			if value > max {
				max = value
			}
		}
	}
}

type channelGauge struct {
	ch chan<- float64
}

func (g *channelGauge) Inc() { g.ch <- +1 }
func (g *channelGauge) Dec() { g.ch <- -1 }
