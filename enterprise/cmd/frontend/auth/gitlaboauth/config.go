package gitlaboauth

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	ffIsEnabled bool
)

const PkgName = "gitlaboauth"

func init() {
	// HACK: don't run this watch loop in tests, because it results in a race condition.
	// This can be removed once the feature flag is removed.
	if strings.HasSuffix(os.Args[0], ".test") {
		ffIsEnabled = true
		return
	}

	go func() {
		conf.Watch(func() {
			if !conf.Get().ExperimentalFeatures.GitlabAuth {
				auth.UpdateProviders(PkgName, nil)
				return
			}

			newProviders, _ := parseConfig(conf.Get())
			if len(newProviders) == 0 {
				auth.UpdateProviders(PkgName, nil)
			} else {
				newProvidersList := make([]auth.Provider, 0, len(newProviders))
				for _, p := range newProviders {
					newProvidersList = append(newProvidersList, p)
				}
				auth.UpdateProviders(PkgName, newProvidersList)
			}
			ffIsEnabled = true
		})
		conf.ContributeValidator(func(cfg conf.Unified) (problems []string) {
			_, problems = parseConfig(&cfg)
			return problems
		})
	}()
}

func parseConfig(cfg *conf.Unified) (providers map[schema.GitLabAuthProvider]auth.Provider, problems []string) {
	providers = make(map[schema.GitLabAuthProvider]auth.Provider)
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
			providers[*pr.Gitlab] = provider
		}
	}
	return providers, problems
}
