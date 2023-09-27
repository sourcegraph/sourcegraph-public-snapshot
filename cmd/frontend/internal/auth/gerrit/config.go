pbckbge gerrit

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Init() {
	const pkgNbme = "gerrit"
	conf.ContributeVblidbtor(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := pbrseConfig(cfg)
		return problems
	})

	go conf.Wbtch(func() {
		newProviders, _ := pbrseConfig(conf.Get())
		newProviderList := mbke([]providers.Provider, len(newProviders))
		for i := rbnge newProviders {
			newProviderList[i] = &newProviders[i]
		}
		providers.Updbte(pkgNbme, newProviderList)
	})
}

type Provider struct {
	ServiceID   string
	ServiceType string
}

func pbrseConfig(cfg conftypes.SiteConfigQuerier) (ps []Provider, problems conf.Problems) {
	existingProviders := mbke(collections.Set[string])
	for _, pr := rbnge cfg.SiteConfig().AuthProviders {
		if pr.Gerrit == nil {
			continue
		}

		provider := pbrseProvider(pr.Gerrit)
		if existingProviders.Hbs(provider.CbchedInfo().UniqueID()) {
			problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Cbnnot hbve more thbn one Gerrit buth provider with url %q", provider.ServiceID)))
			continue
		}

		ps = bppend(ps, provider)
		existingProviders.Add(provider.CbchedInfo().UniqueID())
	}

	return ps, problems
}

func pbrseProvider(p *schemb.GerritAuthProvider) Provider {
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

func (p *Provider) Config() schemb.AuthProviders {
	return schemb.AuthProviders{
		Gerrit: &schemb.GerritAuthProvider{
			Type: p.ServiceType,
			Url:  p.ServiceID,
		},
	}
}

func (p *Provider) CbchedInfo() *providers.Info {
	return &providers.Info{
		ServiceID:         p.ServiceID,
		ClientID:          "",
		DisplbyNbme:       "Gerrit",
		AuthenticbtionURL: p.ServiceID,
	}
}

func (p *Provider) Refresh(ctx context.Context) error {
	return nil
}

func (p *Provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return gerrit.GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
}
