package openidconnect

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func getProviders() []auth.Provider {
	var cfgs []*schema.OpenIDConnectAuthProvider
	for _, p := range conf.Get().Critical.AuthProviders {
		if p.Openidconnect == nil {
			continue
		}
		cfgs = append(cfgs, p.Openidconnect)
	}
	providers := make([]auth.Provider, 0, len(cfgs))
	for _, cfg := range cfgs {
		p := &provider{config: *cfg}
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
						log15.Error("Error prefetching OpenID Connect service provider metadata.", "error", err)
					}
				}(p)
			}
			auth.UpdateProviders("openidconnect", providers)
		})
	}()
}
