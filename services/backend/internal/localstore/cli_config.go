package localstore

import (
	"log"
	"sync"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
)

import srcstore "sourcegraph.com/sourcegraph/srclib/store"

var (
	Defs         = &defs{}
	GlobalDeps   = &globalDeps{}
	GlobalRefs   = &globalRefs{}
	Graph        srcstore.MultiRepoStoreImporterIndexer
	Queue        = &instrumentedQueue{}
	RepoConfigs  = &repoConfigs{}
	RepoStatuses = &repoStatuses{}
	RepoVCS      = &repoVCS{}
	Repos        = &repos{}
)

func init() {
	once := sync.Once{}
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		// initBackground inside of serverctx.Funcs to ensure cli
		// options have already been set.
		once.Do(func() {
			err := initBackground()
			if err != nil {
				log.Fatal(err)
			}
		})
		return ctx, nil
	})

}

// initBackground starts up background store helpers
func initBackground() error {
	c := newQueueStatsCollector(context.Background())
	prometheus.MustRegister(c)

	return nil
}
