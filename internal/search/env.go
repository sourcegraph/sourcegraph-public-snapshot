package search

import (
	"context"
	"sync"

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
		logger := log.Scoped("searcherGRPCConnectionCache")
		searcherGRPCConnectionCache = defaults.NewConnectionCache(logger)
	})

	return searcherGRPCConnectionCache
}

func Indexed() zoekt.Streamer {
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

// ZoektAllIndexed is the subset of zoekt.RepoList that we set in
// ListAllIndexed.
type ZoektAllIndexed struct {
	ReposMap zoekt.ReposMap
	Crashes  int
	Stats    zoekt.RepoStats
}

// ListAllIndexed lists all indexed repositories.
func ListAllIndexed(ctx context.Context, zs zoekt.Searcher) (*ZoektAllIndexed, error) {
	q := &query.Const{Value: true}
	opts := &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMap}

	repos, err := zs.List(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	return &ZoektAllIndexed{
		ReposMap: repos.ReposMap,
		Crashes:  repos.Crashes,
		Stats:    repos.Stats,
	}, nil
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

func reposAtEndpoint(dial func(string) zoekt.Streamer) func(context.Context, string) zoekt.ReposMap {
	return func(ctx context.Context, endpoint string) zoekt.ReposMap {
		cl := dial(endpoint)

		resp, err := cl.List(ctx, &query.Const{Value: true}, &zoekt.ListOptions{Field: zoekt.RepoListFieldReposMap})
		if err != nil {
			return zoekt.ReposMap{}
		}

		return resp.ReposMap
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
