package codeintel

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// InitLSIFStore initializes and returns an LSIF store instance.
func InitLSIFStore() (*lsifstore.Store, error) {
	conn, err := initLSFIStore.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*lsifstore.Store), err
}

var initLSFIStore = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}

	return lsifstore.NewStore(db, conf.Get(), observationContext), nil
})
