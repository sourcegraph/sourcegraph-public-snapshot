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

	mu    sync.RWMutex
	cache map[string]*listCacheValue
}

func NewCachedSearcher(z zoekt.Streamer) zoekt.Streamer {
	return &cachedSearcher{Streamer: z}
}

type listCacheKey struct {
	q    zoektquery.Q
	opts *zoekt.ListOptions
}

func (l listCacheKey) String() string {
	return l.q.String() + " " + l.opts.String()
}

type listCacheValue struct {
	list *zoekt.RepoList
	err  error
	ts   time.Time
}

func (v *listCacheValue) stale() bool {
	return time.Since(v.ts) >= randInterval(5*time.Second, 2*time.Second)
}

func (c *cachedSearcher) String() string {
	return fmt.Sprintf("cachedSearcher(%v)", c.Streamer)
}

func (c *cachedSearcher) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	k := listCacheKey{q: q, opts: opts}

	c.mu.RLock()
	v := c.cache[k.String()]
	c.mu.RUnlock()

	switch {
	case v == nil || v.err != nil:
		c.update(ctx, k) // no cached value, block.
	case v.stale():
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			c.update(ctx, k) // start async update, return stale version
			cancel()
		}()
	}

	return v.list, v.err
}

func (c *cachedSearcher) update(ctx context.Context, k listCacheKey) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	list, err := c.Streamer.List(ctx, k.q, k.opts)
	cancel()

	v := &listCacheValue{
		list: list,
		err:  err,
		ts:   time.Now().UTC(),
	}

	c.mu.Lock()
	c.cache[k.String()] = v
	c.mu.Unlock()
}

// randInterval returns an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randInterval(d, jitter time.Duration) time.Duration {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	return d + delta
}
