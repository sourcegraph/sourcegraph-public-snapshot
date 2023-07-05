package enqueuer

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type operations struct {
	queueIndex           *observation.Operation
	queueIndexForPackage *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_autoindexing_enqueuer",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.enqueuer.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if errors.As(err, &inference.LimitError{}) {
					return observation.EmitForNone
				}
				return observation.EmitForDefault
			},
		})
	}

	return &operations{
		queueIndex:           op("QueueIndex"),
		queueIndexForPackage: op("QueueIndexForPackage"),
	}
}
