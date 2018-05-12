package openidconnect

import (
	"reflect"
	"sync"

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
		init  = true

		mu sync.Mutex
		pc *schema.OpenIDConnectAuthProvider
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		newPC := conf.AuthProvider().Openidconnect
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if first && !init {
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
			first = false
		}
		pc = newPC
		if pc != nil {
			go func(pc schema.OpenIDConnectAuthProvider) {
				if _, err := cache.get(pc.Issuer); err != nil {
					log15.Error("Error prefetching OpenID Connect provider metadata.", "issuer", pc.Issuer, "clientID", pc.ClientID, "error", err)
				}
			}(*pc)
		}
	})
	init = false
}
