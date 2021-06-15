package repos

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var (
	metadataCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repo_metadata_cache_hits",
		Help: "Number of repo metadata cache accesses that were served by the cache",
	})

	metadataCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repo_metadata_cache_misses",
		Help: "Number of repo metadata cache accesses that were not served by the cache and were loaded by the DB",
	})
)

// MetadataCache is an in-memory cache for the metadata of the most frequently searched
// repos. Its intent is to reduce database roundtrips and load by storing the results of
// converting types.RepoName into *types.Repo. With repo results returned in priority order
// by stars, this should have a fairly high hit rate.
// If we move querying the full repo metadata from the database in advance, this will be obsolete.
type MetadataCache struct {
	lru *lru.TwoQueueCache
}

type cacheEntry struct {
	repo       *types.Repo
	validUntil time.Time
}

// Get returns a cached *types.Repo for an id if it exists in the cache
// and hasn't expired.
func (c *MetadataCache) Get(id api.RepoID) (*types.Repo, bool) {
	ci, ok := c.lru.Get(id)
	if !ok {
		metadataCacheMisses.Inc()
		return nil, false
	}
	ce := ci.(cacheEntry)

	// Evict the entry if it has expired
	if time.Now().After(ce.validUntil) {
		c.lru.Remove(id)
		metadataCacheMisses.Inc()
		return nil, false
	}

	metadataCacheHits.Inc()
	return ce.repo, true
}

func (c *MetadataCache) Add(id api.RepoID, repo *types.Repo) {
	e := cacheEntry{
		repo:       repo,
		validUntil: time.Now().Add(10 * time.Minute),
	}
	c.lru.Add(id, e)
}

func NewMetadataCache(size int) MetadataCache {
	l, _ := lru.New2Q(size)
	return MetadataCache{
		lru: l,
	}
}
