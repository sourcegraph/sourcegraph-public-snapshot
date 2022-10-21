package search

import (
	"context"
	"sync"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	indexedSearchOnce sync.Once
	indexedSearch     zoekt.Streamer

	indexersOnce sync.Once
	indexers     *backend.Indexers

	indexedDialerOnce sync.Once
	indexedDialer     backend.ZoektDialer
)

type confEndpointMap struct {
	getter func() []string
}

func (m confEndpointMap) Endpoints() ([]string, error) {
	return m.getter(), nil
}

func (m confEndpointMap) Get(key string) (string, error) {
	em := endpoint.Static(m.getter()...)
	return em.Get(key)
}
func (m confEndpointMap) GetN(key string, n int) ([]string, error) {
	em := endpoint.Static(m.getter()...)
	return em.GetN(key, n)
}

var ErrIndexDisabled = errors.New("indexed search has been disabled")

func SearcherURLs() endpoint.MapLike {
	return confEndpointMap{
		getter: func() []string { return conf.Get().ServiceConnections().Searchers },
	}
}

func Indexed() zoekt.Streamer {
	if !conf.SearchIndexEnabled() {
		return &backend.FakeSearcher{SearchError: ErrIndexDisabled, ListError: ErrIndexDisabled}
	}

	indexedSearchOnce.Do(func() {
		if eps := conf.Get().ServiceConnections().Zoekts; eps != nil {
			indexedSearch = backend.NewCachedSearcher(conf.Get().ServiceConnections().ZoektListTTL, backend.NewMeteredSearcher(
				"", // no hostname means its the aggregator
				&backend.HorizontalSearcher{
					Map: confEndpointMap{
						getter: func() []string { return conf.Get().ServiceConnections().Zoekts },
					},
					Dial: getIndexedDialer(),
				}))
		}
	})

	return indexedSearch
}

// ListAllIndexed lists all indexed repositories with `Minimal: true`. If
// indexed search is disabled it returns ErrIndexDisabled.
func ListAllIndexed(ctx context.Context) (*zoekt.RepoList, error) {
	q := &query.Const{Value: true}
	opts := &zoekt.ListOptions{Minimal: true}
	return Indexed().List(ctx, q, opts)
}

func Indexers() *backend.Indexers {
	indexersOnce.Do(func() {
		indexers = &backend.Indexers{
			Map: confEndpointMap{
				getter: func() []string { return conf.Get().ServiceConnections().Zoekts },
			},
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

		return resp.Minimal
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
