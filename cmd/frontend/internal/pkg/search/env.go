package search

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/google/zoekt/query"
	"github.com/google/zoekt/rpc"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
)

var (
	zoektAddr   = env.Get("ZOEKT_HOST", "indexed-search:80", "host:port of the zoekt instance")
	zoektAddrs  = env.Get("INDEXED_SEARCH_SERVERS", "", "zoekt instances")
	searcherURL = env.Get("SEARCHER_URL", "k8s+http://searcher:3181", "searcher server URL")

	searcherURLsOnce sync.Once
	searcherURLs     *endpoint.Map

	indexedSearchOnce sync.Once
	indexedSearch     *backend.Zoekt

	indexersOnce sync.Once
	indexers     *backend.Indexers
)

func SearcherURLs() *endpoint.Map {
	searcherURLsOnce.Do(func() {
		if len(strings.Fields(searcherURL)) == 0 {
			searcherURLs = endpoint.Empty(errors.New("a searcher service has not been configured"))
		} else {
			searcherURLs = endpoint.New(searcherURL)
		}
	})
	return searcherURLs
}

func Indexed() *backend.Zoekt {
	indexedSearchOnce.Do(func() {
		indexedSearch = &backend.Zoekt{}
		if indexers := Indexers(); indexers.Enabled() {
			indexedSearch.Client = &backend.HorizontalSearcher{
				Map:  indexers.Map,
				Dial: rpc.Client,
			}
		} else if zoektAddr != "" {
			indexedSearch.Client = rpc.Client(zoektAddr)
		}
		conf.Watch(func() {
			indexedSearch.SetEnabled(conf.SearchIndexEnabled())
		})
	})
	return indexedSearch
}

func Indexers() *backend.Indexers {
	indexersOnce.Do(func() {
		if zoektAddrs != "" {
			indexers = &backend.Indexers{
				Map:     endpoint.New(zoektAddrs),
				Indexed: reposAtEndpoint,
			}
		} else {
			indexers = &backend.Indexers{
				Map: nil,
			}
		}
	})
	return indexers
}

func reposAtEndpoint(ctx context.Context, endpoint string) map[string]struct{} {
	cl := rpc.Client(endpoint)
	defer cl.Close()

	resp, err := cl.List(ctx, &query.Const{Value: true})
	if err != nil {
		return map[string]struct{}{}
	}

	set := make(map[string]struct{}, len(resp.Repos))
	for _, r := range resp.Repos {
		set[r.Repository.Name] = struct{}{}
	}
	return set
}
