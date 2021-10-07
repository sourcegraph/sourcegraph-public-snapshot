package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	queuedCount             *observation.Operation
	dequeue                 *observation.Operation
	requeue                 *observation.Operation
	addExecutionLogEntry    *observation.Operation
	updateExecutionLogEntry *observation.Operation
	markComplete            *observation.Operation
	markErrored             *observation.Operation
	markFailed              *observation.Operation
	resetStalled            *observation.Operation
	heartbeat               *observation.Operation
}

func newOperations(storeName string, observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		fmt.Sprintf("workerutil_dbworker_store_%s", storeName),
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(opName string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("workerutil.dbworker.store.%s.%s", storeName, opName),
			MetricLabelValues: []string{opName},
			Metrics:           metrics,
		})
	}

	return &operations{
		queuedCount:             op("QueuedCount"),
		dequeue:                 op("Dequeue"),
		requeue:                 op("Requeue"),
		addExecutionLogEntry:    op("AddExecutionLogEntry"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
		markComplete:            op("MarkComplete"),
		markErrored:             op("MarkErrored"),
		markFailed:              op("MarkFailed"),
		resetStalled:            op("ResetStalled"),
		heartbeat:               op("Heartbeat"),
	}
}
