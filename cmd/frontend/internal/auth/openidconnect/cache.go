package openidconnect

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// cache is the singleton OpenID Connect provider metadata cache.
var cache providerCache

// providerCache caches OpenID Connect provider metadata (which is retrieved from the "issuer
// URL" configured for the provider in site config).
type providerCache struct {
	mu   sync.Mutex
	data map[string]*providerCacheEntry // issuer URL -> entry
}

type providerCacheEntry struct {
	once    sync.Once
	val     *provider
	err     error
	expires time.Time
}

// get gets the OpenID Connect provider at the specified issuer URL. If the provider is cached, it
// returns it from the cache; otherwise it performs a network request to look up the provider. At
// most one network request will be in flight for a given issuerURL; later requests block on the
// original request.
func (c *providerCache) get(issuerURL string) (*provider, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[string]*providerCacheEntry{}
	}
	e, ok := c.data[issuerURL]
	if !ok || time.Now().After(e.expires) {
		e = &providerCacheEntry{}
		c.data[issuerURL] = e
	}
	c.mu.Unlock()

	fetched := false // whether it was fetched in *this* func call
	e.once.Do(func() {
		e.val, e.err = newProvider(context.Background(), issuerURL)
		e.err = errors.WithMessage(e.err, "retrieving OpenID Connect metadata from issuer")
		fetched = true

		var ttl time.Duration
		if e.err == nil {
			ttl = 5 * time.Minute
		} else {
			ttl = 5 * time.Second
		}
		e.expires = time.Now().Add(ttl)
	})

	err := e.err
	if !fetched {
		err = errors.WithMessage(err, "(cached error)") // make debugging easier
	}

	return e.val, err
}
