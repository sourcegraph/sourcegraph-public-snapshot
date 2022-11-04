package graphql

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	documents "github.com/sourcegraph/sourcegraph/internal/codeintel/documents"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	resolver     *Resolver
	resolverOnce sync.Once
)

func GetResolver(svc *documents.Service) *Resolver {
	resolverOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("documents.transport.graphql", "codeintel documents graphql transport"),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}

		resolver = newResolver(svc, observationContext)
	})

	return resolver
}
