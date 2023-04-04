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

type PipelineOptions struct {
	Name        string
	Description string
	Interval    time.Duration
	Metrics     *PipelineMetrics
	ProcessFunc ProcessFunc
}
type ProcessFunc func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered TaggedCounts, err error)

type TaggedCounts interface {
	RecordsAltered() map[string]int
}

type PipelineMetrics struct {
	op                  *observation.Operation
	numRecordsProcessed prometheus.Counter
	numRecordsAltered   *prometheus.CounterVec
}

func NewPipelineMetrics(
	observationCtx *observation.Context,
	name string,
	recordTypeName string,
) *PipelineMetrics {
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

	counterVec := func(name, help string) *prometheus.CounterVec {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, []string{"record"})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numRecordsProcessed := counter(
		fmt.Sprintf("src_%s_records_processed_total", metricName),
		fmt.Sprintf("The number of %s records processed by %s.", recordTypeName, name),
	)

	numRecordsAltered := counterVec(
		fmt.Sprintf("src_%s_records_altered_total", metricName),
		fmt.Sprintf("The number of %s records written/modified by %s.", recordTypeName, name),
	)

	return &PipelineMetrics{
		op:                  op("Handle"),
		numRecordsProcessed: numRecordsProcessed,
		numRecordsAltered:   numRecordsAltered,
	}
}

func NewPipelineJob(ctx context.Context, opts PipelineOptions) goroutine.BackgroundRoutine {
	pipeline := &pipeline{opts: opts}

	return goroutine.NewPeriodicGoroutineWithMetricsAndDynamicInterval(
		ctx,
		opts.Name,
		opts.Description,
		pipeline.interval,
		pipeline,
		opts.Metrics.op,
	)
}

type pipeline struct {
	opts PipelineOptions
	// TODO - metrics about last run to change duration?
}

func (j *pipeline) interval() time.Duration {
	return j.opts.Interval
}

func (j *pipeline) Handle(ctx context.Context) error {
	numRecordsProcessed, numRecordsAltered, err := j.opts.ProcessFunc(ctx)
	if err != nil {
		return err
	}

	j.opts.Metrics.numRecordsProcessed.Add(float64(numRecordsProcessed))

	for name, count := range numRecordsAltered.RecordsAltered() {
		j.opts.Metrics.numRecordsAltered.With(prometheus.Labels{"record": name}).Add(float64(count))
	}

	return nil
}

//
//

type mapCount struct{ value map[string]int }

func (sc mapCount) RecordsAltered() map[string]int { return sc.value }

func NewSingleCount(value int) TaggedCounts {
	return NewMapCount(map[string]int{"record": value})
}

func NewMapCount(value map[string]int) TaggedCounts {
	return &mapCount{value}
}
