package enqueuer

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	queueIndex           *observation.Operation
	queueIndexForPackage *observation.Operation
}

var m = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		r,
		"codeintel_autoindexing_enqueuer",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	m, _ := m.Init(observationContext.Registerer)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.enqueuer.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		queueIndex:           op("QueueIndex"),
		queueIndexForPackage: op("QueueIndexForPackage"),
	}
}
