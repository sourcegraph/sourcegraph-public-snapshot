package saml

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
	var cfgs []*schema.SAMLAuthProvider
	for _, p := range conf.Get().AuthProviders {
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
		const pkgName = "saml"
		logger := log.Scoped(pkgName, "SAML config watch")
		conf.Watch(func() {
			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Warn("Check license for SSO (SAML)", log.Error(err))
				providers.Update(pkgName, nil)
				return
			}

			ps := getProviders()
			for _, p := range ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Background()); err != nil {
						log15.Error("Error prefetching SAML service provider metadata.", "error", err)
					}
				}(p)
			}
			providers.Update(pkgName, ps)
		})
	}()
}
