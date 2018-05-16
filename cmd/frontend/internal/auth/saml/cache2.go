package saml

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/schema"
)

// cache2 is the singleton SAML service provider metadata cache.
var cache2 providerCache2

// providerCache2 caches SAML service provider metadata (which is retrieved using the site config for
// the provider).
type providerCache2 struct {
	mu   sync.Mutex
	data map[schema.SAMLAuthProvider]*providerCacheEntry2 // auth provider config -> entry
}

type providerCacheEntry2 struct {
	once    sync.Once
	val     *saml2.SAMLServiceProvider
	err     error
	expires time.Time
}

var (
	cache2TTLOK  = 5 * time.Minute
	cache2TTLErr = 5 * time.Second
)

// get gets the SAML service provider with the specified config. If the service provider is cached,
// it returns it from the cache; otherwise it performs a network request to look up the provider. At
// most one network request will be in flight for a given provider config; later requests block on
// the original request.
func (c *providerCache2) get(pc schema.SAMLAuthProvider) (*saml2.SAMLServiceProvider, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[schema.SAMLAuthProvider]*providerCacheEntry2{}
	}
	e, ok := c.data[pc]
	if !ok || (!e.expires.IsZero() && time.Now().After(e.expires)) {
		e = &providerCacheEntry2{}
		c.data[pc] = e
	}
	c.mu.Unlock()

	fetched := false // whether it was fetched in *this* func call
	e.once.Do(func() {
		e.val, e.err = getServiceProvider2(context.Background(), &pc)
		e.err = errors.WithMessage(e.err, "retrieving SAML SP metadata from issuer")
		fetched = true

		var ttl time.Duration
		if e.err == nil {
			ttl = cache2TTLOK
		} else {
			ttl = cache2TTLErr
		}
		c.mu.Lock()
		e.expires = time.Now().Add(ttl)
		c.mu.Unlock()
	})

	err := e.err
	if !fetched {
		err = errors.WithMessage(err, "(cached error)") // make debugging easier
	}

	return e.val, err
}
