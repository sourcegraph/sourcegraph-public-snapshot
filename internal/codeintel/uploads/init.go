package uploads

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized uploads service. If the service is
// new, it will use the given database handle.
func GetService(db, codeIntelDB database.DB, gsc GitserverClient) *Service {
	svcOnce.Do(func() {
		logger := log.Scoped(
			"uploads",
			"codeintel uploads service",
		)

		observationContext := &observation.Context{
			Logger:     logger,
			Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
			Registerer: prometheus.DefaultRegisterer,
		}

		store := store.New(db, observationContext)
		locker := locker.NewWith(db, "codeintel")
		repoStore := backend.NewRepos(logger, db)
		lsifstore := lsifstore.New(codeIntelDB, observationContext)

		svc = newService(store, repoStore, lsifstore, gsc, locker, observationContext)
	})

	return svc
}
