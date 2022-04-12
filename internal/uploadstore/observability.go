package uploadstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	Get     *observation.Operation
	Upload  *observation.Operation
	Compose *observation.Operation
	Delete  *observation.Operation
}

func NewOperations(observationContext *observation.Context, domain, storeName string) *Operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		fmt.Sprintf("%s_%s", domain, storeName),
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("%s.%s.%s", domain, storeName, name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		Get:     op("Get"),
		Upload:  op("Upload"),
		Compose: op("Compose"),
		Delete:  op("Delete"),
	}
}
