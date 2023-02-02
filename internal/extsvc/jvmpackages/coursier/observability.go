package coursier

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	log.Logger

	fetchSources  *observation.Operation
	exists        *observation.Operation
	fetchByteCode *observation.Operation
	runCommand    *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_coursier",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.coursier.%s", name),
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
		fetchSources:  op("FetchSources"),
		exists:        op("Exists"),
		fetchByteCode: op("FetchByteCode"),
		runCommand:    op("RunCommand"),

		Logger: observationCtx.Logger,
	}
}

var (
	ops     *operations
	opsOnce sync.Once
)

func getOperations() *operations {
	opsOnce.Do(func() {
		observationCtx := observation.NewContext(log.Scoped("jvmpackages.coursier", ""))

		ops = newOperations(observationCtx)
	})

	return ops
}
