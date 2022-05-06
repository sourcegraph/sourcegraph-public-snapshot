package graphql

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	resolver     *Resolver
	resolverOnce sync.Once
)

func GetResolver(svc *uploads.Service) *Resolver {
	resolverOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("uploads.transport.graphql", "codeintel uploads graphql transport"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		resolver = newResolver(svc, observationContext)
	})

	return resolver
}
