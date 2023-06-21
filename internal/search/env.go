package search

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
)

var (
	searcherURLsOnce sync.Once
	searcherURLs     *endpoint.Map

	searcherGRPCConnectionCacheOnce sync.Once
	searcherGRPCConnectionCache     *defaults.ConnectionCache

	indexedSearchOnce sync.Once
	indexedSearch     zoekt.Streamer

	indexersOnce sync.Once
	indexers     *backend.Indexers

	indexedDialerOnce sync.Once
	indexedDialer     backend.ZoektDialer

	IndexedMock zoekt.Streamer
)

func SearcherURLs() *endpoint.Map {
	searcherURLsOnce.Do(func() {
		searcherURLs = endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
			return conns.Searchers
		})
	})
	return searcherURLs
}

func SearcherGRPCConnectionCache() *defaults.ConnectionCache {
	searcherGRPCConnectionCacheOnce.Do(func() {
		logger := log.Scoped("searcherGRPCConnectionCache", "gRPC connection cache for searcher endpoints")
		searcherGRPCConnectionCache = defaults.NewConnectionCache(logger)
	})

	return searcherGRPCConnectionCache
}

func Indexed() zoekt.Streamer {
	if IndexedMock != nil {
		return IndexedMock
	}
	indexedSearchOnce.Do(func() {
		indexedSearch = backend.NewCachedSearcher(conf.Get().ServiceConnections().ZoektListTTL, backend.NewMeteredSearcher(
			"", // no hostname means its the aggregator
			&backend.HorizontalSearcher{
				Map: endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
					return conns.Zoekts
				}),
				Dial: getIndexedDialer(),
			}))
	})

	return indexedSearch
}

var (
	metricReposLen = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_temp_frontend_index_repos_len",
		Help: "A temporary metric recording different ways to calculate the indexed number of repos.",
	})
	metricReposRepos = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_temp_frontend_index_repos_repos",
		Help: "A temporary metric recording different ways to calculate the indexed number of repos.",
	})
	metricReposCrash = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_temp_frontend_index_repos_crash",
		Help: "A temporary metric recording different ways to calculate the indexed number of repos.",
	})
)

// ListAllIndexed lists all indexed repositories with `Minimal: true`.
func ListAllIndexed(ctx context.Context) (*zoekt.RepoList, error) {
	q := &query.Const{Value: true}
	opts := &zoekt.ListOptions{Minimal: true}

	repos, err := Indexed().List(ctx, q, opts)

	// TODO(keegan) remove this before 2023-08-01. Temporary metric collection.
	if err == nil {
		metricReposLen.Set(float64(len(repos.Minimal))) //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
		metricReposRepos.Set(float64(repos.Stats.Repos))
		metricReposCrash.Set(float64(repos.Crashes))
	}

	return repos, err
}

func Indexers() *backend.Indexers {
	indexersOnce.Do(func() {
		indexers = &backend.Indexers{
			Map: endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
				return conns.Zoekts
			}),
			Indexed: reposAtEndpoint(getIndexedDialer()),
		}
	})
	return indexers
}

func reposAtEndpoint(dial func(string) zoekt.Streamer) func(context.Context, string) map[uint32]*zoekt.MinimalRepoListEntry {
	return func(ctx context.Context, endpoint string) map[uint32]*zoekt.MinimalRepoListEntry {
		cl := dial(endpoint)

		resp, err := cl.List(ctx, &query.Const{Value: true}, &zoekt.ListOptions{Minimal: true})
		if err != nil {
			return map[uint32]*zoekt.MinimalRepoListEntry{}
		}

		return resp.Minimal //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
	}
}

func getIndexedDialer() backend.ZoektDialer {
	indexedDialerOnce.Do(func() {
		indexedDialer = backend.NewCachedZoektDialer(func(endpoint string) zoekt.Streamer {
			return backend.NewCachedSearcher(conf.Get().ServiceConnections().ZoektListTTL, backend.ZoektDial(endpoint))
		})
	})
	return indexedDialer
}
