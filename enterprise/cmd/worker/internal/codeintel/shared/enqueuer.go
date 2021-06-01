package shared

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var initIndexEnqueuerMemo struct {
	conn *enqueuer.IndexEnqueuer
	err  error
	once sync.Once
}

func InitIndexEnqueuer() (*enqueuer.IndexEnqueuer, error) {
	initIndexEnqueuerMemo.once.Do(func() {
		initIndexEnqueuerMemo.conn, initIndexEnqueuerMemo.err = initIndexEnqueuer()
	})

	return initIndexEnqueuerMemo.conn, initIndexEnqueuerMemo.err
}

func initIndexEnqueuer() (*enqueuer.IndexEnqueuer, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}
	gitserverClietn, err := InitGitserverClient()
	if err != nil {
		return nil, err
	}

	return enqueuer.NewIndexEnqueuer(
		&enqueuer.DBStoreShim{Store: dbStore},
		gitserverClietn,
		repoupdater.DefaultClient,
		observationContext,
	), nil
}
