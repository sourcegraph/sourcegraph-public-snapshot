package object

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	Get           *observation.Operation
	Upload        *observation.Operation
	Compose       *observation.Operation
	Delete        *observation.Operation
	ExpireObjects *observation.Operation
	List          *observation.Operation
}

func NewOperations(observationCtx *observation.Context, domain, storeName string) *Operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		fmt.Sprintf("%s_%s", domain, storeName),
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("%s.%s.%s", domain, storeName, name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &Operations{
		Get:           op("Get"),
		Upload:        op("Upload"),
		Compose:       op("Compose"),
		Delete:        op("Delete"),
		ExpireObjects: op("ExpireObjects"),
		List:          op("List"),
	}
}
