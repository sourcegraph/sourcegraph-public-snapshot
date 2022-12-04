package search

import (
	"context"
	"sync"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	searcherURLsOnce sync.Once
	searcherURLs     *endpoint.Map

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
		searcherURLs = endpoint.ConfBased(func(conns conftypes.ServiceConnections) ([]string, bool) {
			return conns.Searchers, false
		})
	})
	return searcherURLs
}

func Indexed() zoekt.Streamer {
	if IndexedMock != nil {
		return IndexedMock
	}
	indexedSearchOnce.Do(func() {
		indexedSearch = backend.NewCachedSearcher(conf.Get().ServiceConnections().ZoektListTTL, backend.NewMeteredSearcher(
			"", // no hostname means its the aggregator
			&backend.HorizontalSearcher{
				Map: endpoint.ConfBased(func(conns conftypes.ServiceConnections) (endpoints []string, intentionallyEmpty bool) {
					return conns.Zoekts, conns.ZoektsIntentionallyEmpty
				}),
				Dial: getIndexedDialer(),
			}))
	})

	return indexedSearch
}

// ListAllIndexed lists all indexed repositories with `Minimal: true`. If any
// crashes occur an error is returned instead of returning partial results.
func ListAllIndexed(ctx context.Context) (*zoekt.RepoList, error) {
	q := &query.Const{Value: true}
	opts := &zoekt.ListOptions{Minimal: true}
	rl, err := Indexed().List(ctx, q, opts)
	if err != nil {
		return nil, err
	}

	if rl.Crashes > 0 {
		return nil, errors.New("zoekt.List call occurred while not all Zoekt replicas are available")
	}

	return rl, nil
}

func Indexers() *backend.Indexers {
	indexersOnce.Do(func() {
		indexers = &backend.Indexers{
			Map: endpoint.ConfBased(func(conns conftypes.ServiceConnections) (endpoints []string, intentionallyEmpty bool) {
				return conns.Zoekts, conns.ZoektsIntentionallyEmpty
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
