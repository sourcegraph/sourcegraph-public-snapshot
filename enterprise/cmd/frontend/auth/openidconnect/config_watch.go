package openidconnect

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of issuer metadata (given the configured OpenID Connect issuer
// URL) immediately upon server startup and site config changes so users don't incur the wait on the
// first auth flow request.
func init() {
	providersOfType := func(ps []schema.AuthProviders) []*schema.OpenIDConnectAuthProvider {
		var pcs []*schema.OpenIDConnectAuthProvider
		for _, p := range ps {
			if p.Openidconnect != nil {
				pcs = append(pcs, p.Openidconnect)
			}
		}
		return pcs
	}

	var (
		init = true

		mu  sync.Mutex
		cur []*schema.OpenIDConnectAuthProvider
		reg = map[schema.OpenIDConnectAuthProvider]auth.Provider{}
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		new := providersOfType(conf.Get().AuthProviders)
		diff := diffProviderConfig(cur, new)
		if len(diff) == 0 {
			return
		}

		if !init {
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
		}
		updates := make(map[auth.Provider]bool, len(diff))
		for pc, op := range diff {
			if old, ok := reg[pc]; ok {
				delete(reg, pc)
				updates[old] = false
			}
			if op {
				new := &provider{config: pc}
				reg[pc] = new
				updates[new] = true
				go func(p *provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching OpenID Connect service provider metadata.", "error", err)
					}
				}(new)
			}
		}
		auth.UpdateProviders(updates)
		cur = new
	})
	init = false
}

func diffProviderConfig(old, new []*schema.OpenIDConnectAuthProvider) map[schema.OpenIDConnectAuthProvider]bool {
	diff := map[schema.OpenIDConnectAuthProvider]bool{}
	for _, oldPC := range old {
		diff[*oldPC] = false
	}
	for _, newPC := range new {
		if _, ok := diff[*newPC]; ok {
			delete(diff, *newPC)
		} else {
			diff[*newPC] = true
		}
	}
	return diff
}
