// Package zoekt provides a client to github.com/google/zoekt
package zoekt

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
)

// req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
// 	nethttp.OperationName("Zoekt "+strings.ToTitle(method)),
// 	nethttp.ClientTrace(false))
// defer ht.Finish()

var listQueryAll = &query.Const{Value: true}

// Cache caches responses that rarely change from Zoekt. Note: Cache starts up
// background goroutines, so call Stop when done using the Client.
type Cache struct {
	Client zoekt.Searcher

	mu       sync.Mutex
	state    int32 // 0 not running, 1 running, 2 stopped
	listResp *zoekt.RepoList
	listErr  error
}

// Stop will tear down the background goroutines.
func (c *Cache) Stop() {
	c.mu.Lock()
	c.state = 2
	c.mu.Unlock()
}

// ListAll returns the response of List without any restrictions.
func (c *Cache) ListAll(ctx context.Context) (*zoekt.RepoList, error) {
	c.mu.Lock()
	r, err := c.listResp, c.listErr
	c.mu.Unlock()

	// No cached responses, start up and just do uncached query.
	if r == nil && err == nil {
		go c.start()
		r, err = c.Client.List(ctx, listQueryAll)
	}

	return r, err
}

func (c *Cache) start() {
	c.mu.Lock()
	if c.state != 0 {
		// already running or stopped
		c.mu.Unlock()
		return
	}
	c.state = 1 // mark running
	c.mu.Unlock()

	errorCount := 0
	state := int32(1)
	for state == 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		listResp, listErr := c.Client.List(ctx, listQueryAll)
		cancel()

		// Only update on error once it has happened 3 times in a row. This is
		// to prevent us caching transient errors, and instead fallback on the
		// old list.
		update := true
		if listErr != nil {
			errorCount++
			if errorCount <= 3 {
				update = false
			}
		} else {
			errorCount = 0
		}

		c.mu.Lock()
		state = c.state
		if update {
			c.listResp, c.listErr = listResp, listErr
		}
		c.mu.Unlock()

		randSleep(5*time.Second, 2*time.Second)
	}
}

// randSleep will sleep for an expected d duration with a jitter in [-jitter /
// 2, jitter / 2].
func randSleep(d, jitter time.Duration) {
	delta := time.Duration(rand.Int63n(int64(jitter))) - (jitter / 2)
	time.Sleep(d + delta)
}
