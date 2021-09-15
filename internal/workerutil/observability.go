package workerutil

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type WorkerMetrics struct {
	operations *operations
	numJobs    Gauge
}

type Gauge interface {
	Inc()
	Dec()
}

type operations struct {
	handle *observation.Operation
}

// NewMetrics creates and registers the following metrics for a generic worker instance.
//
//   - {prefix}_duration_seconds_bucket: handler operation latency histogram
//   - {prefix}_total: number of handler operations
//   - {prefix}_error_total: number of handler operations resulting in an error
//   - {prefix}_handlers: the number of active handler routines
//
// The given labels are emitted on each metric.
func NewMetrics(observationContext *observation.Context, prefix string, labels map[string]string) WorkerMetrics {
	keys := make([]string, 0, len(labels))
	values := make([]string, 0, len(labels))
	for key, value := range labels {
		keys = append(keys, key)
		values = append(values, value)
	}

	gauge := func(name, help string) prometheus.Gauge {
		gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("src_%s_%s", prefix, name),
			Help: help,
		}, keys)

		observationContext.Registerer.MustRegister(gaugeVec)
		return gaugeVec.WithLabelValues(values...)
	}

	numJobs := gauge(
		"handlers",
		"The number of active handlers.",
	)

	return WorkerMetrics{
		operations: newOperations(observationContext, prefix, keys, values),
		numJobs:    newLenientConcurrencyGauge(numJobs, time.Second*5),
	}
}

func newOperations(observationContext *observation.Context, prefix string, keys, values []string) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		prefix,
		metrics.WithLabels(append(keys, "op")...),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("worker.%s", name),
			MetricLabelValues: append(append([]string{}, values...), name),
			Metrics:           metrics,
		})
	}

	return &operations{
		handle: op("Handle"),
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
