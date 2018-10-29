package saml

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of SAML IdP metadata immediately upon server startup and site
// config changes so users don't incur the wait on the first auth flow request.
func init() {
	providersOfType := func(ps []schema.AuthProviders) []*schema.SAMLAuthProvider {
		var pcs []*schema.SAMLAuthProvider
		for _, p := range ps {
			if p.Saml != nil {
				pcs = append(pcs, withConfigDefaults(p.Saml))
			}
		}
		return pcs
	}

	var (
		init = true

		mu  sync.Mutex
		cur []*schema.SAMLAuthProvider
		reg = map[schema.SAMLAuthProvider]auth.Provider{}
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
			log15.Info("Reloading changed SAML authentication provider configuration.")
		}
		multiple := len(new) >= 2
		updates := make(map[auth.Provider]bool, len(diff))
		for pc, op := range diff {
			if old, ok := reg[pc]; ok {
				delete(reg, pc)
				updates[old] = false
			}
			if op {
				new := &provider{config: pc, multiple: multiple}
				reg[pc] = new
				updates[new] = true
				go func(p *provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching SAML service provider metadata.", "error", err)
					}
				}(new)
			}
		}
		auth.UpdateProviders(updates)
		cur = new
	})
	init = false
}

func diffProviderConfig(old, new []*schema.SAMLAuthProvider) map[schema.SAMLAuthProvider]bool {
	diff := map[schema.SAMLAuthProvider]bool{}
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
