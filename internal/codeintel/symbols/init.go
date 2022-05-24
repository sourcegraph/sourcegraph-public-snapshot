package symbols

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized symbols service. If the service is
// new, it will use the given database handle.
func GetService(db database.DB) *Service {
	svcOnce.Do(func() {
		storeObservationCtx := &observation.Context{
			Logger:     log.Scoped("symbols.store", "codeintels symbols store"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		store := store.New(db, storeObservationCtx)

		observationCtx := &observation.Context{
			Logger:     log.Scoped("symbols.service", "codeintel symbols service"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(store, observationCtx)
	})

	return svc
}
