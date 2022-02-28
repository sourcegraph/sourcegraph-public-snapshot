package dependencies

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	depSvc     *Service
	depSvcOnce sync.Once
)

func GetService(db database.DB, syncer Syncer) *Service {
	depSvcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		depSvc = newService(
			store.GetStore(db),
			lockfiles.GetService(
				authz.DefaultSubRepoPermsChecker,
				git.LsFiles,
				gitserver.DefaultClient.Archive,
			),
			syncer,
			observationContext,
		)
	})

	return depSvc
}
