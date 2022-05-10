package store

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	store     *Store
	storeOnce sync.Once
)

func GetStore(db database.DB) *Store {
	storeOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("documents.store", "codeintel documents store"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		store = newStore(db, observationContext)
	})

	return store
}

func TestStore(db database.DB) *Store {
	return newStore(db, &observation.TestContext)
}
