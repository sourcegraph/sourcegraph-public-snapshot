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
	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		_, messages := parseConfig(&cfg)
		return conf.NewCriticalProblems(messages...)
	})
	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(conf.Get())
			if len(newProviders) == 0 {
				providers.Update(PkgName, nil)
			} else {
				newProvidersList := make([]providers.Provider, 0, len(newProviders))
				for _, p := range newProviders {
					newProvidersList = append(newProvidersList, p)
				}
				providers.Update(PkgName, newProvidersList)
			}
		})
	}()
}

func parseConfig(cfg *conf.Unified) (ps map[schema.GitLabAuthProvider]providers.Provider, messages []string) {
	ps = make(map[schema.GitLabAuthProvider]providers.Provider)
	for _, pr := range cfg.Critical.AuthProviders {
		if pr.Gitlab == nil {
			continue
		}

		if cfg.Critical.ExternalURL == "" {
			messages = append(messages, "`externalURL` was empty and it is needed to determine the OAuth callback URL.")
			continue
		}
		externalURL, err := url.Parse(cfg.Critical.ExternalURL)
		if err != nil {
			messages = append(messages, fmt.Sprintf("Could not parse `externalURL`, which is needed to determine the OAuth callback URL."))
			continue
		}
		callbackURL := *externalURL
		callbackURL.Path = "/.auth/gitlab/callback"

		provider, providerMessages := parseProvider(callbackURL.String(), pr.Gitlab, pr)
		messages = append(messages, providerMessages...)
		if provider != nil {
			ps[*pr.Gitlab] = provider
		}
	}
	return ps, messages
}
