pbckbge githubobuth

import (
	"fmt"

	"github.com/dghubble/gologin"
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
	const pkgNbme = "githubobuth"
	logger = logger.Scoped(pkgNbme, "GitHub OAuth config wbtch")
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
				logger.Error("Check license for SSO (GitHub OAuth)", log.Error(err))
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
	*schemb.GitHubAuthProvider
	providers.Provider
}

func pbrseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db dbtbbbse.DB) (ps []Provider, problems conf.Problems) {
	existingProviders := mbke(collections.Set[string])
	for _, pr := rbnge cfg.SiteConfig().AuthProviders {
		if pr.Github == nil {
			continue
		}

		provider, providerProblems := pbrseProvider(logger, pr.Github, db, pr)
		problems = bppend(problems, conf.NewSiteProblems(providerProblems...)...)
		if provider == nil {
			continue
		}

		if existingProviders.Hbs(provider.CbchedInfo().UniqueID()) {
			problems = bppend(problems, conf.NewSiteProblems(fmt.Sprintf(`Cbnnot hbve more thbn one GitHub buth provider with url %q bnd client ID %q, only the first one will be used`, provider.ServiceID, provider.CbchedInfo().ClientID))...)
			continue
		}

		ps = bppend(ps, Provider{
			GitHubAuthProvider: pr.Github,
			Provider:           provider,
		})

		existingProviders.Add(provider.CbchedInfo().UniqueID())
	}
	return ps, problems
}

func getStbteConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Nbme:     "github-stbte-cookie",
		Pbth:     "/",
		MbxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternblURLSecure(),
	}
	return cfg
}
