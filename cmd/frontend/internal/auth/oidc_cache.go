package auth

import (
	"context"
	"reflect"
	"sync"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of issuer metadata (given the configured OpenID Connect issuer
// URL) immediately upon server startup and site config changes so users don't incur the wait on the
// first auth flow request.
func init() {
	var (
		first = true

		oidcMu sync.Mutex
		pc     *schema.OpenIDConnectAuthProvider
	)
	conf.Watch(func() {
		oidcMu.Lock()
		defer oidcMu.Unlock()

		// Only react when the config changes.
		newPC := conf.AuthProvider().Openidconnect
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if first {
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
			first = false
		}
		pc = newPC
		if pc != nil {
			go func(pc schema.OpenIDConnectAuthProvider) {
				if _, err := oidcCache.get(pc.Issuer); err != nil {
					log15.Error("Error prefetching OpenID Connect provider metadata.", "issuer", pc.Issuer, "clientID", pc.ClientID, "error", err)
				}
			}(*pc)
		}
	})
}

// oidcCache is the singleton OpenID Connect provider metadata cache.
var oidcCache oidcProviderCache

// oidcProviderCache caches OpenID Connect provider metadata (which is retrieved from the "issuer
// URL" configured for the provider in site config).
type oidcProviderCache struct {
	mu   sync.Mutex
	data map[string]*oidcProviderCacheEntry // issuer URL -> entry
}

type oidcProviderCacheEntry struct {
	once    sync.Once
	val     *oidc.Provider
	err     error
	expires time.Time
}

// get gets the OpenID Connect provider at the specified issuer URL. If the provider is cached, it
// returns it from the cache; otherwise it performs a network request to look up the provider. At
// most one network request will be in flight for a given issuerURL; later requests block on the
// original request.
func (c *oidcProviderCache) get(issuerURL string) (*oidc.Provider, error) {
	c.mu.Lock()
	if c.data == nil {
		c.data = map[string]*oidcProviderCacheEntry{}
	}
	e, ok := c.data[issuerURL]
	if !ok || time.Now().After(e.expires) {
		e = &oidcProviderCacheEntry{}
		c.data[issuerURL] = e
	}
	c.mu.Unlock()

	fetched := false // whether it was fetched in *this* func call
	e.once.Do(func() {
		e.val, e.err = oidc.NewProvider(context.Background(), issuerURL)
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
