package githuboauth

import (
	"github.com/dghubble/gologin"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init(db database.DB) {
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(cfg, db)
		return problems
	})

	const pkgName = "githuboauth"
	logger := log.Scoped(pkgName, "GitHub OAuth config watch")
	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(conf.Get(), db)
			if len(newProviders) == 0 {
				providers.Update(pkgName, nil)
				return
			}

			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Error("Check license for SSO (GitHub OAuth)", log.Error(err))
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
	*schema.GitHubAuthProvider
	providers.Provider
}

func parseConfig(cfg conftypes.SiteConfigQuerier, db database.DB) (ps []Provider, problems conf.Problems) {
	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.Github == nil {
			continue
		}

		provider, providerProblems := parseProvider(pr.Github, db, pr)
		problems = append(problems, conf.NewSiteProblems(providerProblems...)...)
		if provider != nil {
			ps = append(ps, Provider{
				GitHubAuthProvider: pr.Github,
				Provider:           provider,
			})
		}
	}
	return ps, problems
}

func getStateConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     "github-state-cookie",
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternalURLSecure(),
	}
	return cfg
}
