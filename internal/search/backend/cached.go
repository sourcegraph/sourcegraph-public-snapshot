package backend

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
)

// cachedSearcher wraps a zoekt.Searcher with caching of List call results.
type cachedSearcher struct {
	zoekt.Streamer

	ttl   time.Duration
	now   func() time.Time
	mu    sync.RWMutex
	cache map[listCacheKey]*listCacheValue
}

func NewCachedSearcher(ttl time.Duration, z zoekt.Streamer) *cachedSearcher {
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
	if v, ok := q.(*zoektquery.Const); !ok || !v.Value {
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

	c.cache[k] = v

	return v
}

// randInterval returns an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randInterval(d, jitter time.Duration) time.Duration {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	return d + delta
}
