package codeintel

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// InitDBStore initializes and returns a db store instance.
func InitDBStore() (*dbstore.Store, error) {
	conn, err := initDBStore.Init()
	return conn.(*dbstore.Store), err
}

var initDBStore = shared.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := shared.InitDatabase()
	if err != nil {
		return nil, err
	}

	return dbstore.NewWithDB(db, observationContext), nil
})
