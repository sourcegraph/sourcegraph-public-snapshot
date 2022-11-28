package npm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	fetchSources *observation.Operation
	exists       *observation.Operation
	runCommand   *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_npm",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.npm.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if err != nil && strings.Contains(err.Error(), "not found") {
					return observation.EmitForMetrics | observation.EmitForTraces
				}
				return observation.EmitForDefault
			},
		})
	}

	return &operations{
		fetchSources: op("FetchSources"),
		exists:       op("Exists"),
		runCommand:   op("RunCommand"),
	}
}

var (
	ops     *operations
	opsOnce sync.Once
)

func getOperations() *operations {
	opsOnce.Do(func() {
		observationContext := observation.NewContext(log.Scoped("npm", ""))

		ops = newOperations(observationContext)
	})

	return ops
}
