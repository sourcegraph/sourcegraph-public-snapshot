package gerrit

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init() {
	const pkgName = "gerrit"
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(cfg)
		return problems
	})

	go conf.Watch(func() {
		newProviders, _ := parseConfig(conf.Get())
		newProviderList := make([]providers.Provider, len(newProviders))
		for i := range newProviders {
			newProviderList[i] = &newProviders[i]
		}
		providers.Update(pkgName, newProviderList)
	})
}

type Provider struct {
	ServiceID   string
	ServiceType string
}

func parseConfig(cfg conftypes.SiteConfigQuerier) (ps []Provider, problems conf.Problems) {
	existingProviders := make(collections.Set[string])
	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.Gerrit == nil {
			continue
		}

		provider := parseProvider(pr.Gerrit)
		if existingProviders.Has(provider.CachedInfo().UniqueID()) {
			problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Cannot have more than one Gerrit auth provider with url %q", provider.ServiceID)))
			continue
		}

		ps = append(ps, provider)
		existingProviders.Add(provider.CachedInfo().UniqueID())
	}

	return ps, problems
}

func parseProvider(p *schema.GerritAuthProvider) Provider {
	return Provider{
		ServiceID:   p.Url,
		ServiceType: p.Type,
	}
}

func (p *Provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: extsvc.TypeGerrit,
		ID:   p.ServiceID,
	}
}

func (p *Provider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Gerrit: &schema.GerritAuthProvider{
			Type: p.ServiceType,
			Url:  p.ServiceID,
		},
	}
}

func (p *Provider) CachedInfo() *providers.Info {
	return &providers.Info{
		ServiceID:         p.ServiceID,
		ClientID:          "",
		DisplayName:       "Gerrit",
		AuthenticationURL: p.ServiceID,
	}
}

func (p *Provider) Refresh(ctx context.Context) error {
	return nil
}

func (p *Provider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	return gerrit.GetPublicExternalAccountData(ctx, &account.AccountData)
}
