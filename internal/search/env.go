package search

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/google/zoekt"
	"github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
)

var (
	searcherURL = env.Get("SEARCHER_URL", "k8s+http://searcher:3181", "searcher server URL")

	searcherURLsOnce sync.Once
	searcherURLs     *endpoint.Map

	indexedEndpointsOnce sync.Once
	indexedEndpoints     *endpoint.Map

	indexedSearchOnce sync.Once
	indexedSearch     zoekt.Streamer

	indexersOnce sync.Once
	indexers     *backend.Indexers

	indexedDialer = backend.NewCachedZoektDialer(backend.ZoektDial)
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

func IndexedEndpoints() *endpoint.Map {
	indexedEndpointsOnce.Do(func() {
		if addr := zoektAddr(os.Environ()); addr != "" {
			indexedEndpoints = endpoint.New(addr)
		}
	})
	return indexedEndpoints
}

func Indexed() zoekt.Streamer {
	indexedSearchOnce.Do(func() {
		if eps := IndexedEndpoints(); eps != nil {
			indexedSearch = backend.NewCachedSearcher(backend.NewMeteredSearcher(
				"", // no hostname means its the aggregator
				&backend.HorizontalSearcher{
					Map:  eps,
					Dial: indexedDialer,
				}))
		}
	})
	return indexedSearch
}

func Indexers() *backend.Indexers {
	indexersOnce.Do(func() {
		indexers = &backend.Indexers{
			Map:     IndexedEndpoints(),
			Indexed: reposAtEndpoint(indexedDialer),
		}
	})
	return indexers
}

func zoektAddr(environ []string) string {
	if addr, ok := getEnv(environ, "INDEXED_SEARCH_SERVERS"); ok {
		return addr
	}

	// Backwards compatibility: We used to call this variable ZOEKT_HOST
	if addr, ok := getEnv(environ, "ZOEKT_HOST"); ok {
		return addr
	}

	// Not set, use the default (service discovery on the indexed-search
	// statefulset)
	return "k8s+rpc://indexed-search:6070?kind=sts"
}

func getEnv(environ []string, key string) (string, bool) {
	key = key + "="
	for _, env := range environ {
		if strings.HasPrefix(env, key) {
			return env[len(key):], true
		}
	}
	return "", false
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
