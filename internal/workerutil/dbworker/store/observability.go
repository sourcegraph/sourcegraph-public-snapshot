package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	addExecutionLogEntry    *observation.Operation
	dequeue                 *observation.Operation
	heartbeat               *observation.Operation
	markComplete            *observation.Operation
	markErrored             *observation.Operation
	markFailed              *observation.Operation
	maxDurationInQueue      *observation.Operation
	queuedCount             *observation.Operation
	requeue                 *observation.Operation
	resetStalled            *observation.Operation
	updateExecutionLogEntry *observation.Operation
}

func newOperations(storeName string, observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
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
		addExecutionLogEntry:    op("AddExecutionLogEntry"),
		dequeue:                 op("Dequeue"),
		heartbeat:               op("Heartbeat"),
		markComplete:            op("MarkComplete"),
		markErrored:             op("MarkErrored"),
		markFailed:              op("MarkFailed"),
		maxDurationInQueue:      op("MaxDurationInQueue"),
		queuedCount:             op("QueuedCount"),
		requeue:                 op("Requeue"),
		resetStalled:            op("ResetStalled"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
	}
}
