package codeintel

import (
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// InitLSIFStore initializes and returns an LSIF store instance.
func InitLSIFStore() (*lsifstore.Store, error) {
	return initLSFIStore.Init()
}

var initLSFIStore = memo.NewMemoizedConstructor(func() (*lsifstore.Store, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("store.lsif", "lsif store"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}

	return lsifstore.NewStore(stores.NewCodeIntelDB(db), conf.Get(), observationContext), nil
})
