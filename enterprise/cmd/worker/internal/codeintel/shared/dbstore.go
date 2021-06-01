package shared

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	enterpriseshared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var initDBStoreMemo struct {
	conn *dbstore.Store
	err  error
	once sync.Once
}

func InitDBStore() (*dbstore.Store, error) {
	initDBStoreMemo.once.Do(func() {
		initDBStoreMemo.conn, initDBStoreMemo.err = initDBStore()
	})

	return initDBStoreMemo.conn, initDBStoreMemo.err
}

func initDBStore() (*dbstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := enterpriseshared.InitDatabase()
	if err != nil {
		return nil, err
	}

	return dbstore.NewWithDB(db, observationContext), nil
}
