package codenav

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized symbols service. If the service is
// new, it will use the given database handle.
func GetService(db, codeIntelDB database.DB, uploadSvc UploadService, gitserver GitserverClient) *Service {
	svcOnce.Do(func() {
		oc := func(name string) *observation.Context {
			return &observation.Context{
				Logger:     log.Scoped("symbols."+name, "codeintel symbols "+name),
				Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
				Registerer: prometheus.DefaultRegisterer,
			}
		}

		store := store.New(db, oc("store"))
		lsifstore := lsifstore.New(codeIntelDB, oc("lsifstore"))
		svc = newService(store, lsifstore, uploadSvc, gitserver, oc("service"))
	})

	return svc
}
