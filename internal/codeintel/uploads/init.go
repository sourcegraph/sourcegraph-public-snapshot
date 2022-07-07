package uploads

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
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
func GetService(db, codeIntelDB database.DB, gsc shared.GitserverClient) *Service {
	svcOnce.Do(func() {
		lg := func(name string) log.Logger {
			return log.Scoped("uploads."+name, "codeintel uploads "+name)
		}

		oc := func(name string) *observation.Context {
			return &observation.Context{
				Logger:     lg(name),
				Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
				Registerer: prometheus.DefaultRegisterer,
			}
		}

		lsifstore := lsifstore.New(codeIntelDB, oc("lsifstore"))
		store := store.New(db, oc("store"))
		locker := locker.NewWith(db, "codeintel")

		svc = newService(store, lsifstore, gsc, locker, oc("service"))
	})

	return svc
}

// Need it specifically for connecting to the lsifstore database.
func mustInitializeCodeIntelDB(logger log.Logger) stores.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	db, err := connections.EnsureNewCodeIntelDB(dsn, "codeintel", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to codeintel database", log.Error(err))
	}
	return stores.NewCodeIntelDB(db)
}
