package graphql

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	autoindexing "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	resolver     *Resolver
	resolverOnce sync.Once
)

func GetResolver(svc *autoindexing.Service) *Resolver {
	resolverOnce.Do(func() {
		observationContext := &observation.Context{
			Logger: log.Scoped("autoindexing.transport.graphql", "codeintel autoindexing graphql transport"),
			Tracer: &trace.Tracer{TracerProvider: otel.GetTracerProvider()},

			Registerer: prometheus.DefaultRegisterer,
		}

		resolver = newResolver(svc, observationContext)
	})

	return resolver
}
