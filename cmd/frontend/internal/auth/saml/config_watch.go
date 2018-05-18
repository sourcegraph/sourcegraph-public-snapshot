package saml

import (
	"context"
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type configOp int

const (
	opAdded configOp = iota
	opChanged
	opRemoved
)

// Start trying to populate the cache of SAML IdP metadata immediately upon server startup and site
// config changes so users don't incur the wait on the first auth flow request.
func init() {
	providersOfType := func(ps []schema.AuthProviders) []*schema.SAMLAuthProvider {
		var pcs []*schema.SAMLAuthProvider
		for _, p := range ps {
			if p.Saml != nil {
				pcs = append(pcs, p.Saml)
			}
		}
		return pcs
	}

	var (
		init = true

		mu  sync.Mutex
		cur []*schema.SAMLAuthProvider
		reg = map[providerID]auth.Provider{}
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		new := providersOfType(conf.AuthProviders())
		diff := diffProviderConfig(cur, new)
		if len(diff) == 0 {
			return
		}

		if !init {
			log15.Info("Reloading changed SAML authentication provider configuration.")
		}
		updates := make(map[auth.Provider]bool, len(diff))
		for pc, op := range diff {
			pcKey := toProviderID(&pc)
			if old, ok := reg[pcKey]; ok {
				delete(reg, pcKey)
				updates[old] = false
			}
			if op == opAdded || op == opChanged {
				new := &provider{config: pc}
				reg[pcKey] = new
				updates[new] = true

				go func(p *provider) {
					var err error
					if conf.EnhancedSAMLEnabled() {
						err = p.Refresh(context.Background())
					} else {
						_, err = cache1.get(pc)
					}
					if err != nil {
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

func toKeyMap(pcs []*schema.SAMLAuthProvider) map[providerID]*schema.SAMLAuthProvider {
	m := make(map[providerID]*schema.SAMLAuthProvider, len(pcs))
	for _, pc := range pcs {
		m[toProviderID(pc)] = pc
	}
	return m
}

func diffProviderConfig(old, new []*schema.SAMLAuthProvider) map[schema.SAMLAuthProvider]configOp {
	oldMap := toKeyMap(old)
	diff := map[schema.SAMLAuthProvider]configOp{}
	for _, newPC := range new {
		newKey := toProviderID(newPC)
		if oldPC, ok := oldMap[newKey]; ok {
			if !reflect.DeepEqual(oldPC, newPC) {
				diff[*newPC] = opChanged
			}
			delete(oldMap, newKey)
		} else {
			diff[*newPC] = opAdded
		}
	}
	for _, oldPC := range oldMap {
		diff[*oldPC] = opRemoved
	}
	return diff
}
