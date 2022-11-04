package dependencies

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized dependencies service. If the service is
// new, it will use the given database handle and git/syncer instances.
func GetService(db database.DB) *Service {
	svcOnce.Do(func() {
		storeObservationCtx := &observation.Context{
			Logger:     log.Scoped("dependencies.store", "dependencies store"),
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}
		var store store.Store = store.New(db, storeObservationCtx)

		svc = newService(store)
	})

	return svc
}

// TestService creates a fresh dependencies service with the given database handle and git/syncer
// instances.
func TestService(db database.DB) *Service {
	store := store.New(db, &observation.TestContext)

	return newService(store)
}
