package autoindexing

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized autoindexing service. If the service is
// new, it will use the given database handle.

func GetService(db database.DB) *Service {
	svcOnce.Do(func() {
		storeObservationCtx := &observation.Context{
			Logger:     log.Scoped("autoindexing.store", "autoindexing store"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		store := store.New(db, storeObservationCtx)

		observationCxt := &observation.Context{
			Logger:     log.Scoped("autoindexing.service", "autoindexing service"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(store, observationCxt)
	})

	return svc
}

// To be removed after https://github.com/sourcegraph/sourcegraph/issues/33377

type InferenceService = inference.Service

var GetInferenceService = inference.GetService
