package graphql

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	documents "github.com/sourcegraph/sourcegraph/internal/codeintel/documents"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	resolver     *Resolver
	resolverOnce sync.Once
)

func GetResolver(svc *documents.Service) *Resolver {
	resolverOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("documents.transport.graphql", "codeintel documents graphql transport"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		resolver = newResolver(svc, observationContext)
	})

	return resolver
}
