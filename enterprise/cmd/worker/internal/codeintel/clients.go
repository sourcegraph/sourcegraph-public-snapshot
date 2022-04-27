package codeintel

import (
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// InitGitserverClient initializes and returns a gitserver client.
func InitGitserverClient() (*gitserver.Client, error) {
	conn, err := initGitserverClient.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*gitserver.Client), err
}

var initGitserverClient = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("client.gitserver", "gitserver client"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dbStore, err := InitDBStore()
	if err != nil {
		return nil, err
	}

	return gitserver.New(database.NewDBWith(dbStore), dbStore, observationContext), nil
})

func InitRepoUpdaterClient() *repoupdater.Client {
	client, _ := initRepoUpdaterClient.Init()
	return client.(*repoupdater.Client)
}

var initRepoUpdaterClient = memo.NewMemoizedConstructor(func() (interface{}, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("client.repo-updater", "repo-updater client"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return repoupdater.New(observationContext), nil
})
