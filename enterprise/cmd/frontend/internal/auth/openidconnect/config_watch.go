package openidconnect

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func getProviders() []providers.Provider {
	var cfgs []*schema.OpenIDConnectAuthProvider
	for _, p := range conf.Get().AuthProviders {
		if p.Openidconnect == nil {
			continue
		}
		cfgs = append(cfgs, p.Openidconnect)
	}
	ps := make([]providers.Provider, 0, len(cfgs))
	for _, cfg := range cfgs {
		p := &provider{config: *cfg}
		ps = append(ps, p)
	}
	return ps
}

func init() {
	go func() {
		const pkgName = "openidconnect"
		logger := log.Scoped(pkgName, "OpenID Connect config watch")
		conf.Watch(func() {
			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Warn("Check license for SSO (OpenID Connect)", log.Error(err))
				providers.Update(pkgName, nil)
				return
			}

			ps := getProviders()
			for _, p := range ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching OpenID Connect service provider metadata.", "error", err)
					}
				}(p)
			}
			providers.Update(pkgName, ps)
		})
	}()
}
