package shared

import (
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var initGitserverClientMemo struct {
	conn *gitserver.Client
	err  error
	once sync.Once
}

func InitGitserverClient() (*gitserver.Client, error) {
	initGitserverClientMemo.once.Do(func() {
		initGitserverClientMemo.conn, initGitserverClientMemo.err = initGitserverClient()
	})

	return initGitserverClientMemo.conn, initGitserverClientMemo.err
}

func initGitserverClient() (*gitserver.Client, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return gitserver.New(dbStore, observationContext), nil
}
