package shared

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	enterpriseshared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var initLSIFStoreMemo struct {
	conn *lsifstore.Store
	err  error
	once sync.Once
}

func InitLSIFStore() (*lsifstore.Store, error) {
	initLSIFStoreMemo.once.Do(func() {
		initLSIFStoreMemo.conn, initLSIFStoreMemo.err = initLSIFStore()
	})

	return initLSIFStoreMemo.conn, initLSIFStoreMemo.err
}

func initLSIFStore() (*lsifstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := enterpriseshared.InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}

	return lsifstore.NewStore(db, observationContext), nil
}
