package autoindexing

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

var (
	maximumRepositoriesInspectedPerSecond    = toRate(env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_INSPECTED_PER_SECOND", 0, "The maximum number of repositories inspected for auto-indexing per second. Set to zero to disable limit."))
	maximumRepositoriesUpdatedPerSecond      = toRate(env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_UPDATED_PER_SECOND", 0, "The maximum number of repositories cloned or fetched for auto-indexing per second. Set to zero to disable limit."))
	maximumIndexJobsPerInferredConfiguration = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 25, "Repositories with a number of inferred auto-index jobs exceeding this threshold will not be auto-indexed.")
)

// GetService creates or returns an already-initialized autoindexing service. If the service is
// new, it will use the given database handle.
func GetService(db database.DB, uploadSvc shared.UploadService, gitserver shared.GitserverClient, repoUpdater shared.RepoUpdaterClient) *Service {
	svcOnce.Do(func() {
		oc := func(name string) *observation.Context {
			return &observation.Context{
				Logger:     log.Scoped("autoindexing."+name, "autoindexing "+name),
				Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
				Registerer: prometheus.DefaultRegisterer,
			}
		}

		s := store.New(db, oc("store"))
		inf := inference.GetService(db)

		svc = newService(s, uploadSvc, gitserver, repoUpdater, inf, oc("service"))
	})

	return svc
}

func toRate(value int) rate.Limit {
	if value == 0 {
		return rate.Inf
	}

	return rate.Limit(value)
}
