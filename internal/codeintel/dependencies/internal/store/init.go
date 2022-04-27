package store

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	ops     *operations
	opsOnce sync.Once
)

func GetStore(db database.DB) *Store {
	// newOperations registers Prometheus metrics with MustRegister,
	// which panics if we register the same metrics twice, so we protect
	// it with this package level sync.Once.
	opsOnce.Do(func() {
		ops = newOperations(&observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		})
	})
	return newStore(db, ops)
}

func TestStore(db database.DB) *Store {
	return newStore(db, newOperations(&observation.TestContext))
}
