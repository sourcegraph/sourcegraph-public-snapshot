package gitlaboauth

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

const PkgName = "gitlaboauth"

func init() {
	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(conf.Get())
			if len(newProviders) == 0 {
				providers.UpdateProviders(PkgName, nil)
			} else {
				newProvidersList := make([]providers.Provider, 0, len(newProviders))
				for _, p := range newProviders {
					newProvidersList = append(newProvidersList, p)
				}
				providers.UpdateProviders(PkgName, newProvidersList)
			}
		})
		conf.ContributeValidator(func(cfg conf.Unified) (problems []string) {
			_, problems = parseConfig(&cfg)
			return problems
		})
	}()
}

func parseConfig(cfg *conf.Unified) (ps map[schema.GitLabAuthProvider]providers.Provider, problems []string) {
	ps = make(map[schema.GitLabAuthProvider]providers.Provider)
	for _, pr := range cfg.Critical.AuthProviders {
		if pr.Gitlab == nil {
			continue
		}

		if cfg.Critical.ExternalURL == "" {
			problems = append(problems, "`externalURL` was empty and it is needed to determine the OAuth callback URL.")
			continue
		}
		externalURL, err := url.Parse(cfg.Critical.ExternalURL)
		if err != nil {
			problems = append(problems, fmt.Sprintf("Could not parse `externalURL`, which is needed to determine the OAuth callback URL."))
			continue
		}
		callbackURL := *externalURL
		callbackURL.Path = "/.auth/gitlab/callback"

		provider, providerProblems := parseProvider(callbackURL.String(), pr.Gitlab, pr)
		problems = append(problems, providerProblems...)
		if provider != nil {
			ps[*pr.Gitlab] = provider
		}
	}
	return ps, problems
}
