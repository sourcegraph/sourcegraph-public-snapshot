package dependencies

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

func GetService(db database.DB, lockfilesService LockfilesService, syncer Syncer) *Service {
	svcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(
			store.GetStore(db),
			lockfilesService,
			syncer,
			observationContext,
		)
	})

	return svc
}
