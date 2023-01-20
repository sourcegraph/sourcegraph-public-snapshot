package job

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	addExecutionLogEntry    *observation.Operation
	updateExecutionLogEntry *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"apiworker_apiclient_queue_job",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.apiclient.queue.job.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		addExecutionLogEntry:    op("AddExecutionLogEntry"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
	}
}
