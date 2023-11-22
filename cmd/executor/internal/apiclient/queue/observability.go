package queue

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	dequeue                 *observation.Operation
	markComplete            *observation.Operation
	markErrored             *observation.Operation
	markFailed              *observation.Operation
	heartbeat               *observation.Operation
	addExecutionLogEntry    *observation.Operation
	updateExecutionLogEntry *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"apiworker_apiclient_queue",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.apiclient.queue.worker.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		dequeue:                 op("Dequeue"),
		markComplete:            op("MarkComplete"),
		markErrored:             op("MarkErrored"),
		markFailed:              op("MarkFailed"),
		heartbeat:               op("Heartbeat"),
		addExecutionLogEntry:    op("AddExecutionLogEntry"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
	}
}
