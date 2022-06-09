package codeintel

import (
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// InitGitserverClient initializes and returns a gitserver client.
func InitGitserverClient() (*gitserver.Client, error) {
	return initGitserverClient.Init()
}

var initGitserverClient = memo.NewMemoizedConstructor(func() (*gitserver.Client, error) {
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
	return client
}

var initRepoUpdaterClient = memo.NewMemoizedConstructor(func() (*repoupdater.Client, error) {
	observationContext := &observation.Context{
		Logger:     log.Scoped("client.repo-updater", "repo-updater client"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return repoupdater.New(observationContext), nil
})
