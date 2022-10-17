package coursier

import (
	"fmt"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type operations struct {
	log.Logger

	fetchSources  *observation.Operation
	exists        *observation.Operation
	fetchByteCode *observation.Operation
	runCommand    *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_coursier",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.coursier.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
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

		Logger: observationContext.Logger,
	}
}

var (
	ops     *operations
	opsOnce sync.Once
)

func getOperations() *operations {
	opsOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("jvmpackages.coursier", ""),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}

		ops = newOperations(observationContext)
	})

	return ops
}
