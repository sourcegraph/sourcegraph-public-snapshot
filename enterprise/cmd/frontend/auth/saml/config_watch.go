package saml

import (
	"context"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/cmd/frontend/auth/providers"
	"sourcegraph.com/pkg/conf"
	"sourcegraph.com/schema"
)

func getProviders() []providers.Provider {
	var cfgs []*schema.SAMLAuthProvider
	for _, p := range conf.Get().Critical.AuthProviders {
		if p.Saml == nil {
			continue
		}
		cfgs = append(cfgs, withConfigDefaults(p.Saml))
	}
	multiple := len(cfgs) >= 2
	ps := make([]providers.Provider, 0, len(cfgs))
	for _, cfg := range cfgs {
		p := &provider{config: *cfg, multiple: multiple}
		ps = append(ps, p)
	}
	return ps
}

func init() {
	go func() {
		conf.Watch(func() {
			ps := getProviders()
			for _, p := range ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching SAML service provider metadata.", "error", err)
					}
				}(p)
			}
			providers.Update("saml", ps)
		})
	}()
}
