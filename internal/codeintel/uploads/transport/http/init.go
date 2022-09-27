package http

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	handler     http.Handler
	handlerOnce sync.Once
)

func GetHandler(svc *uploads.Service, withCodeHostAuthAuth bool) http.Handler {
	handlerOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("uploads.handler", "codeintel uploads http handler"),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}

		handler = newHandler(svc, observationContext)
	})

	return handler
}
