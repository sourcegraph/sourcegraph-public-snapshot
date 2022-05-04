package graphql

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	symbols "github.com/sourcegraph/sourcegraph/internal/codeintel/symbols"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	resolver     *Resolver
	resolverOnce sync.Once
)

func GetResolver(svc *symbols.Service) *Resolver {
	resolverOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("symbols.transport.graphql", "codeintel symbols graphql transport"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		resolver = newResolver(svc, observationContext)
	})

	return resolver
}
