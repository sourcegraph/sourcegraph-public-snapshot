package policies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	list        *observation.Operation
	get         *observation.Operation
	create      *observation.Operation
	update      *observation.Operation
	delete      *observation.Operation
	findMatches *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_policies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		list:        op("List"),
		get:         op("Get"),
		create:      op("Create"),
		update:      op("Update"),
		delete:      op("Delete"),
		findMatches: op("FindMatches"),
	}
}
