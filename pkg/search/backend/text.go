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

// Zoekt wraps a zoekt.Searcher.
//
// Note: Zoekt starts up background goroutines, so call Close when done using
// the Client.
type Zoekt struct {
	Client zoekt.Searcher

	// DisableCache when true prevents caching of Client.List. Useful in
	// tests.
	DisableCache bool

	mu       sync.Mutex
	state    int32 // 0 not running, 1 running, 2 stopped
	listResp *zoekt.RepoList
	listErr  error
	disabled bool
}

// Close will tear down the background goroutines.
func (c *Zoekt) Close() {
	c.mu.Lock()
	c.state = 2
	c.mu.Unlock()
}

func (c *Zoekt) String() string {
	return fmt.Sprintf("zoekt(%v)", c.Client)
}

// ListAll returns the response of List without any restrictions.
func (c *Zoekt) ListAll(ctx context.Context) (*zoekt.RepoList, error) {
	if !c.Enabled() {
		// By returning an empty list Text.Search won't send any queries to
		// Zoekt.
		return &zoekt.RepoList{}, nil
	}

	c.mu.Lock()
	r, err := c.listResp, c.listErr
	c.mu.Unlock()

	// No cached responses, start up and just do uncached query.
	if r == nil && err == nil {
		if !c.DisableCache {
			go c.start()
		}
		r, err = c.Client.List(ctx, &zoektquery.Const{Value: true})
	}

	return r, err
}

// SetEnabled will disable zoekt if b is false.
func (c *Zoekt) SetEnabled(b bool) {
	c.mu.Lock()
	c.disabled = !b
	c.mu.Unlock()
}

// Enabled returns true if Zoekt is enabled. It is enabled if Client is
// non-nil and it hasn't been disabled by SetEnable.
func (c *Zoekt) Enabled() bool {
	c.mu.Lock()
	b := c.disabled
	c.mu.Unlock()
	return c.Client != nil && !b
}

// start starts a goroutine that keeps the listResp and listErr fields updated
// from the Zoekt server, as a local cache.
func (c *Zoekt) start() {
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
		if !c.Enabled() {
			// If we haven't been stopped, reset state so start() is called
			// again when we are enabled. We can defer unlocking since we will
			// return.
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.state == 1 {
				c.state = 0
				c.listResp, c.listErr = nil, nil
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		listResp, listErr := c.Client.List(ctx, &zoektquery.Const{Value: true})
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
