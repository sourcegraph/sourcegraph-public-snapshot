package autoindexing

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	svc     *Service
	svcOnce sync.Once
)

var (
	maximumRepositoriesInspectedPerSecond    = toRate(env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_INSPECTED_PER_SECOND", 0, "The maximum number of repositories inspected for auto-indexing per second. Set to zero to disable limit."))
	maximumRepositoriesUpdatedPerSecond      = toRate(env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_UPDATED_PER_SECOND", 0, "The maximum number of repositories cloned or fetched for auto-indexing per second. Set to zero to disable limit."))
	maximumIndexJobsPerInferredConfiguration = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 25, "Repositories with a number of inferred auto-index jobs exceeding this threshold will be auto-indexed.")
)

// GetService creates or returns an already-initialized autoindexing service. If the service is
// new, it will use the given database handle.
func GetService(
	db database.DB,
	dbStore DBStore,
	gitserverClient GitserverClient,
	repoUpdater RepoUpdaterClient,
) *Service {
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

		svc = newService(
			store,
			dbStore,
			gitserverClient,
			repoUpdater,
			inference.GetService(db),
			observationCxt,
		)
	})

	return svc
}

func toRate(value int) rate.Limit {
	if value == 0 {
		return rate.Inf
	}

	return rate.Limit(value)
}

// To be removed after https://github.com/sourcegraph/sourcegraph/issues/33377

type InferenceService = inference.Service

var GetInferenceService = inference.GetService
