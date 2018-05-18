package openidconnect

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
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
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
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error refreshing OpenID Connect provider metadata.", "issuer", p.config.Issuer, "clientID", p.config.ClientID, "error", err)
					}
				}(new)
			}
		}
		auth.UpdateProviders(updates)
		cur = new
	})
	init = false
}

func toKeyMap(pcs []*schema.OpenIDConnectAuthProvider) map[providerID]*schema.OpenIDConnectAuthProvider {
	m := make(map[providerID]*schema.OpenIDConnectAuthProvider, len(pcs))
	for _, pc := range pcs {
		m[toProviderID(pc)] = pc
	}
	return m
}

func diffProviderConfig(old, new []*schema.OpenIDConnectAuthProvider) map[schema.OpenIDConnectAuthProvider]configOp {
	oldMap := toKeyMap(old)
	diff := map[schema.OpenIDConnectAuthProvider]configOp{}
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
