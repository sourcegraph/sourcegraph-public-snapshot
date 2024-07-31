package gitlaboauth

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init(logger log.Logger, db database.DB) {
	const pkgName = "gitlaboauth"
	logger = log.Scoped(pkgName)

	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(logger, cfg, db)
		return problems
	})

	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(logger, conf.Get(), db)
			if len(newProviders) == 0 {
				providers.Update(pkgName, nil)
				return
			}

			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Error("Check license for SSO (GitLab OAuth)", log.Error(err))
				providers.Update(pkgName, nil)
				return
			}

			newProvidersList := make([]providers.Provider, 0, len(newProviders))
			for _, p := range newProviders {
				newProvidersList = append(newProvidersList, p.Provider)
			}
			providers.Update(pkgName, newProvidersList)
		})
	}()
}

type Provider struct {
	*schema.GitLabAuthProvider
	providers.Provider
}

func parseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db database.DB) (ps []Provider, problems conf.Problems) {
	existingProviders := make(collections.Set[string])
	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.Gitlab == nil {
			continue
		}

		if cfg.SiteConfig().ExternalURL == "" {
			problems = append(problems, conf.NewSiteProblem("`externalURL` was empty and it is needed to determine the OAuth callback URL."))
			continue
		}
		externalURL, err := url.Parse(cfg.SiteConfig().ExternalURL)
		if err != nil {
			problems = append(problems, conf.NewSiteProblem("Could not parse `externalURL`, which is needed to determine the OAuth callback URL."))
			continue
		}
		callbackURL := *externalURL
		callbackURL.Path = "/.auth/gitlab/callback"

		provider, providerMessages := parseProvider(logger, db, callbackURL.String(), pr.Gitlab, pr)
		problems = append(problems, conf.NewSiteProblems(providerMessages...)...)
		if provider == nil {
			continue
		}

		if existingProviders.Has(provider.CachedInfo().UniqueID()) {
			problems = append(problems, conf.NewSiteProblems(fmt.Sprintf(`Cannot have more than one GitLab auth provider with url %q and client ID %q, only the first one will be used`, provider.ServiceID, provider.CachedInfo().ClientID))...)
			continue
		}

		ps = append(ps, Provider{
			GitLabAuthProvider: pr.Gitlab,
			Provider:           provider,
		})

		existingProviders.Add(provider.CachedInfo().UniqueID())
	}
	return ps, problems
}
