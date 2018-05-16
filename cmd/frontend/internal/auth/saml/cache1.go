package saml

import (
	"context"
	"sync"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// cache1 is the singleton SAML service provider metadata cache.
var cache1 providerCache1

// providerCache1 caches SAML service provider metadata (which is retrieved using the site config for
// the provider).
type providerCache1 struct {
	mu   sync.Mutex
	data map[schema.SAMLAuthProvider]*providerCacheEntry1 // auth provider config -> entry
}

type providerCacheEntry1 struct {
	once    sync.Once
	val     *samlsp.Middleware
	err     error
	expires time.Time
}

// get gets the SAML service provider with the specified config. If the service provider is cached,
// it returns it from the cache; otherwise it performs a network request to look up the provider. At
// most one network request will be in flight for a given provider config; later requests block on
// the original request.
func (c *providerCache1) get(pc schema.SAMLAuthProvider) (*samlsp.Middleware, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[schema.SAMLAuthProvider]*providerCacheEntry1{}
	}
	e, ok := c.data[pc]
	if !ok || (!e.expires.IsZero() && time.Now().After(e.expires)) {
		e = &providerCacheEntry1{}
		c.data[pc] = e
	}
	c.mu.Unlock()

	fetched := false // whether it was fetched in *this* func call
	e.once.Do(func() {
		e.val, e.err = getServiceProvider1(context.Background(), &pc)
		e.err = errors.WithMessage(e.err, "retrieving SAML SP metadata from issuer")
		fetched = true

		var ttl time.Duration
		if e.err == nil {
			ttl = 5 * time.Minute
		} else {
			ttl = 5 * time.Second
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
