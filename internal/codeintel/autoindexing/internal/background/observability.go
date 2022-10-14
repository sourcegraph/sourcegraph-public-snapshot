package background

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type operations struct {
	// Indexes
	queueIndexForPackage *observation.Operation
	queueIndex           *observation.Operation

	handleIndexScheduler *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_background",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.background.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	handleIndexScheduler := observationContext.Operation(observation.Op{
		Name:              "codeintel.indexing.HandleIndexSchedule",
		MetricLabelValues: []string{"HandleIndexSchedule"},
		Metrics:           m,
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			if errors.As(err, &inference.LimitError{}) {
				return observation.EmitForDefault.Without(observation.EmitForMetrics)
			}
			return observation.EmitForDefault
		},
	})

	return &operations{
		// Indexes
		queueIndexForPackage: op("QueueIndexForPackage"),
		queueIndex:           op("QueueIndex"),

		handleIndexScheduler: handleIndexScheduler,
	}
}
