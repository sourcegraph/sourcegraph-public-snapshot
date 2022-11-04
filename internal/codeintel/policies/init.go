package policies

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	svc     *Service
	svcOnce sync.Once
)

// GetService creates or returns an already-initialized policies service. If the service is
// new, it will use the given database handle.
func GetService(db database.DB, uploadSvc UploadService, gitserver GitserverClient) *Service {
	svcOnce.Do(func() {
		oc := func(name string) *observation.Context {
			return &observation.Context{
				Logger:     log.Scoped("policies."+name, "codeintel policies "+name),
				Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
				Registerer: prometheus.DefaultRegisterer,
			}
		}

		store := store.New(db, oc("store"))
		svc = newService(store, uploadSvc, gitserver, oc("service"))
	})

	return svc
}
