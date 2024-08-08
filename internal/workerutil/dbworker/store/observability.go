package store

import (
	"fmt"
	"sync"

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
	countByState            *observation.Operation
	exists                  *observation.Operation
	requeue                 *observation.Operation
	resetStalled            *observation.Operation
	updateExecutionLogEntry *observation.Operation
	canceledJobs            *observation.Operation
}

// as newOperations changes based on the store name passed in, and a dbworker store
// for a given store can be created more than once (once for actual use and once for metrics),
// we avoid a "panic: duplicate metrics collector registration attempted" this way.
var (
	metricsMap = map[string]*metrics.REDMetrics{}
	metricsMu  sync.Mutex
)

func newOperations(observationCtx *observation.Context, storeName string) *operations {
	metricsMu.Lock()

	var red *metrics.REDMetrics
	if m, ok := metricsMap[storeName]; ok {
		red = m
	} else {
		red = metrics.NewREDMetrics(
			observationCtx.Registerer,
			fmt.Sprintf("workerutil_dbworker_store_%s", storeName),
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
		metricsMap[storeName] = red
	}

	metricsMu.Unlock()

	op := func(opName string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("workerutil.dbworker.store.%s.%s", storeName, opName),
			MetricLabelValues: []string{opName},
			Metrics:           red,
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
		countByState:            op("CountByState"),
		exists:                  op("Exists"),
		requeue:                 op("Requeue"),
		resetStalled:            op("ResetStalled"),
		updateExecutionLogEntry: op("UpdateExecutionLogEntry"),
		canceledJobs:            op("CanceledJobs"),
	}
}
