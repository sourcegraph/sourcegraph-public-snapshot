package http

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	handler         http.Handler
	handlerWithAuth http.Handler
	handlerOnce     sync.Once
)

func GetHandler(svc *uploads.Service, db database.DB, withCodeHostAuthAuth bool) http.Handler {
	handlerOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("uploads.handler", "codeintel uploads http handler"),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}

		operations := newOperations(observationContext)
		handler = newHandler(svc, operations)

		// 🚨 SECURITY: Non-internal installations of this handler will require a user/repo
		// visibility check with the remote code host (if enabled via site configuration).
		handlerWithAuth = auth.AuthMiddleware(handler, db, auth.DefaultValidatorByCodeHost, operations.authMiddleware)
	})

	if withCodeHostAuthAuth {
		return handlerWithAuth
	}
	return handler
}
