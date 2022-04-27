package http

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
	handler     *Handler
	handlerOnce sync.Once
)

func GetHandler(svc *uploads.Service) *Handler {
	handlerOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("uploads.handler", "codeintel uploads http handler"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		handler = newHandler(svc, observationContext)
	})

	return handler
}
