package background

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type JanitorOptions struct {
	Name        string
	Description string
	Interval    time.Duration
	Metrics     *JanitorMetrics
	CleanupFunc CleanupFunc
}

type CleanupFunc func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, err error)

type JanitorMetrics struct {
	op                *observation.Operation
	numRecordsScanned prometheus.Counter
	numRecordsAltered prometheus.Counter
}

func NewJanitorMetrics(
	observationCtx *observation.Context,
	redMetrics *metrics.REDMetrics,
	name string,
	recordTypeName string,
) *JanitorMetrics {
	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              name,
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	replacer := strings.NewReplacer(
		".", "_",
		"-", "_",
	)

	numRecordsScanned := counter(
		fmt.Sprintf("src_%s_records_scanned_total", replacer.Replace(name)),
		fmt.Sprintf("The number of %s records scanned by %s.", recordTypeName, name),
	)
	numRecordsAltered := counter(
		fmt.Sprintf("src_%s_records_altered_total", replacer.Replace(name)),
		fmt.Sprintf("The number of %s records altered by %s.", recordTypeName, name),
	)

	return &JanitorMetrics{
		op:                op("Handle"),
		numRecordsScanned: numRecordsScanned,
		numRecordsAltered: numRecordsAltered,
	}
}

func NewJanitorJob(ctx context.Context, opts JanitorOptions) goroutine.BackgroundRoutine {
	janitor := &janitor{opts: opts}

	return goroutine.NewPeriodicGoroutineWithMetricsAndDynamicInterval(
		ctx,
		opts.Name,
		opts.Description,
		janitor.interval,
		janitor,
		opts.Metrics.op,
	)
}

type janitor struct {
	opts JanitorOptions
	// TODO - metrics about last run to change duration?
}

func (j *janitor) interval() time.Duration {
	return j.opts.Interval
}

func (j *janitor) Handle(ctx context.Context) error {
	numRecordsScanned, numRecordsAltered, err := j.opts.CleanupFunc(ctx)
	if err != nil {
		return err
	}

	j.opts.Metrics.numRecordsScanned.Add(float64(numRecordsScanned))
	j.opts.Metrics.numRecordsAltered.Add(float64(numRecordsAltered))
	return nil
}
