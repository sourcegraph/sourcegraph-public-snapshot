package saml

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func getProviders() []auth.Provider {
	var cfgs []*schema.SAMLAuthProvider
	for _, p := range conf.Get().Critical.AuthProviders {
		if p.Saml == nil {
			continue
		}
		cfgs = append(cfgs, withConfigDefaults(p.Saml))
	}
	multiple := len(cfgs) >= 2
	providers := make([]auth.Provider, 0, len(cfgs))
	for _, cfg := range cfgs {
		p := &provider{config: *cfg, multiple: multiple}
		providers = append(providers, p)
	}
	return providers
}

func init() {
	go func() {
		conf.Watch(func() {
			providers := getProviders()
			for _, p := range providers {
				go func(p auth.Provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching SAML service provider metadata.", "error", err)
					}
				}(p)
			}
			auth.UpdateProviders("saml", providers)
		})
	}()
}
