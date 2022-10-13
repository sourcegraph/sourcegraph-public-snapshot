package queue

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	dequeue                 *observation.Operation
	addExecutionLogEntry    *observation.Operation
	updateExecutionLogEntry *observation.Operation
	markComplete            *observation.Operation
	markErrored             *observation.Operation
	markFailed              *observation.Operation
	heartbeat               *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"apiworker_apiclient_queue",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.apiclient.queue.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		dequeue:                 op("Dequeue"),
		addExecutionLogEntry:    op("AddExecutionLogEntry"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
		markComplete:            op("MarkComplete"),
		markErrored:             op("MarkErrored"),
		markFailed:              op("MarkFailed"),
		heartbeat:               op("Heartbeat"),
	}
}
