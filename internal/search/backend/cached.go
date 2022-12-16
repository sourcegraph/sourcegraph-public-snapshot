package backend

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

// cachedSearcher wraps a zoekt.Searcher with caching of List call results.
type cachedSearcher struct {
	zoekt.Streamer

	ttl   time.Duration
	now   func() time.Time
	mu    sync.RWMutex
	cache map[listCacheKey]*listCacheValue
}

func NewCachedSearcher(ttl time.Duration, z zoekt.Streamer) zoekt.Streamer {
	return &cachedSearcher{
		Streamer: z,
		ttl:      ttl,
		now:      time.Now,
		cache:    map[listCacheKey]*listCacheValue{},
	}
}

type listCacheKey struct {
	opts zoekt.ListOptions
}

type listCacheValue struct {
	list *zoekt.RepoList
	err  error
	ts   time.Time
	now  func() time.Time
	ttl  time.Duration
}

func (v *listCacheValue) stale() bool {
	return v.now().Sub(v.ts) >= randInterval(v.ttl, 5*time.Second)
}

func (c *cachedSearcher) String() string {
	return fmt.Sprintf("cachedSearcher(%s, %v)", c.ttl, c.Streamer)
}

func (c *cachedSearcher) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if !isTrueQuery(q) {
		// cache pass-through for anything that isn't "ListAll", either minimal or not
		return c.Streamer.List(ctx, q, opts)
	}

	k := listCacheKey{}
	if opts != nil {
		k.opts = *opts
	}

	c.mu.RLock()
	v := c.cache[k]
	c.mu.RUnlock()

	switch {
	case v == nil || v.err != nil:
		v = c.update(ctx, q, k) // no cached value, block.
	case v.stale():
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			c.update(ctx, q, k) // start async update, return stale version
			cancel()
		}()
	}

	return v.list, v.err
}

// isTrueQuery returns true if q will always match all shards.
func isTrueQuery(q zoektquery.Q) bool {
	// the query is probably wrapped to avoid extra RPC work.
	q = zoektquery.RPCUnwrap(q)

	v, ok := q.(*zoektquery.Const)
	return ok && v.Value
}

func (c *cachedSearcher) update(ctx context.Context, q zoektquery.Q, k listCacheKey) *listCacheValue {
	c.mu.Lock()
	defer c.mu.Unlock()

	v := c.cache[k]
	if v != nil && v.err == nil && !v.stale() {
		// someone beat us to the update
		return v
	}

	list, err := c.Streamer.List(ctx, q, &k.opts)

	v = &listCacheValue{
		list: list,
		err:  err,
		ttl:  c.ttl,
		now:  c.now,
		ts:   c.now(),
	}

	// If we encountered an error or a crash, shorten how long we wait before
	// refreshing the cache.
	if err != nil || list.Crashes > 0 {
		v.ttl /= 4
	}

	c.cache[k] = v

	return v
}

// randInterval returns an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randInterval(d, jitter time.Duration) time.Duration {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	return d + delta
}
