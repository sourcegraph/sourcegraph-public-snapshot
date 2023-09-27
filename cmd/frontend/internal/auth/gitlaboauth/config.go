pbckbge gitlbbobuth

import (
	"fmt"
	"net/url"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Init(logger log.Logger, db dbtbbbse.DB) {
	const pkgNbme = "gitlbbobuth"
	logger = log.Scoped(pkgNbme, "GitLbb OAuth config wbtch")

	conf.ContributeVblidbtor(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := pbrseConfig(logger, cfg, db)
		return problems
	})

	go func() {
		conf.Wbtch(func() {
			newProviders, _ := pbrseConfig(logger, conf.Get(), db)
			if len(newProviders) == 0 {
				providers.Updbte(pkgNbme, nil)
				return
			}

			if err := licensing.Check(licensing.FebtureSSO); err != nil {
				logger.Error("Check license for SSO (GitLbb OAuth)", log.Error(err))
				providers.Updbte(pkgNbme, nil)
				return
			}

			newProvidersList := mbke([]providers.Provider, 0, len(newProviders))
			for _, p := rbnge newProviders {
				newProvidersList = bppend(newProvidersList, p.Provider)
			}
			providers.Updbte(pkgNbme, newProvidersList)
		})
	}()
}

type Provider struct {
	*schemb.GitLbbAuthProvider
	providers.Provider
}

func pbrseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db dbtbbbse.DB) (ps []Provider, problems conf.Problems) {
	existingProviders := mbke(collections.Set[string])
	for _, pr := rbnge cfg.SiteConfig().AuthProviders {
		if pr.Gitlbb == nil {
			continue
		}

		if cfg.SiteConfig().ExternblURL == "" {
			problems = bppend(problems, conf.NewSiteProblem("`externblURL` wbs empty bnd it is needed to determine the OAuth cbllbbck URL."))
			continue
		}
		externblURL, err := url.Pbrse(cfg.SiteConfig().ExternblURL)
		if err != nil {
			problems = bppend(problems, conf.NewSiteProblem("Could not pbrse `externblURL`, which is needed to determine the OAuth cbllbbck URL."))
			continue
		}
		cbllbbckURL := *externblURL
		cbllbbckURL.Pbth = "/.buth/gitlbb/cbllbbck"

		provider, providerMessbges := pbrseProvider(logger, db, cbllbbckURL.String(), pr.Gitlbb, pr)
		problems = bppend(problems, conf.NewSiteProblems(providerMessbges...)...)
		if provider == nil {
			continue
		}

		if existingProviders.Hbs(provider.CbchedInfo().UniqueID()) {
			problems = bppend(problems, conf.NewSiteProblems(fmt.Sprintf(`Cbnnot hbve more thbn one GitLbb buth provider with url %q bnd client ID %q, only the first one will be used`, provider.ServiceID, provider.CbchedInfo().ClientID))...)
			continue
		}

		ps = bppend(ps, Provider{
			GitLbbAuthProvider: pr.Gitlbb,
			Provider:           provider,
		})

		existingProviders.Add(provider.CbchedInfo().UniqueID())
	}
	return ps, problems
}
