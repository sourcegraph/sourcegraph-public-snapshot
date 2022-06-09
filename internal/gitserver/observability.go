package gitserver

import (
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type operations struct {
	batchLog       *observation.Operation
	batchLogSingle *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"gitserver_client",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("gitserver.client.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	// suboperations do not have their own metrics but do have their
	// own opentracing spans. This allows us to more granularly track
	// the latency for parts of a request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name: fmt.Sprintf("gitserver.client.%s", name),
		})
	}

	return &operations{
		batchLog:       op("BatchLog"),
		batchLogSingle: subOp("batchLogSingle"),
	}
}

var (
	operationsInst     *operations
	operationsInstOnce sync.Once
)

func getOperations() *operations {
	operationsInstOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("gitserver.client", "gitserver client"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		operationsInst = newOperations(observationContext)
	})

	return operationsInst
}
