package background

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/actor"
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
	name string,
) *JanitorMetrics {
	replacer := strings.NewReplacer(
		".", "_",
		"-", "_",
	)
	metricName := replacer.Replace(name)

	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		metricName,
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

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

	numRecordsScanned := counter(
		fmt.Sprintf("src_%s_records_scanned_total", metricName),
		fmt.Sprintf("The number of records scanned by %s.", name),
	)
	numRecordsAltered := counter(
		fmt.Sprintf("src_%s_records_altered_total", metricName),
		fmt.Sprintf("The number of records altered by %s.", name),
	)

	return &JanitorMetrics{
		op:                op("Handle"),
		numRecordsScanned: numRecordsScanned,
		numRecordsAltered: numRecordsAltered,
	}
}

func NewJanitorJob(ctx context.Context, opts JanitorOptions) goroutine.BackgroundRoutine {
	janitor := &janitor{opts: opts}

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		janitor,
		goroutine.WithName(opts.Name),
		goroutine.WithDescription(opts.Description),
		goroutine.WithIntervalFunc(janitor.interval),
		goroutine.WithOperation(opts.Metrics.op),
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
