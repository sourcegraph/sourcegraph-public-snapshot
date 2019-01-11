package githuboauth

import (
	"os"
	"strings"

	"github.com/dghubble/gologin"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	ffIsEnabled bool
)

func init() {
	// HACK: don't run this watch loop in tests, because it results in a race condition.
	// This can be removed once the feature flag is removed.
	if strings.HasSuffix(os.Args[0], ".test") {
		ffIsEnabled = true
		return
	}

	const pkgName = "githuboauth"
	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(conf.Get())
			if len(newProviders) == 0 {
				auth.UpdateProviders(pkgName, nil)
			} else {
				newProvidersList := make([]auth.Provider, 0, len(newProviders))
				for _, p := range newProviders {
					newProvidersList = append(newProvidersList, p)
				}
				auth.UpdateProviders(pkgName, newProvidersList)
			}
			ffIsEnabled = true
		})
		conf.ContributeValidator(func(cfg conf.Unified) (problems []string) {
			_, problems = parseConfig(&cfg)
			return problems
		})
	}()
}

func parseConfig(cfg *conf.Unified) (providers map[schema.GitHubAuthProvider]auth.Provider, problems []string) {
	providers = make(map[schema.GitHubAuthProvider]auth.Provider)
	for _, pr := range cfg.Critical.AuthProviders {
		if pr.Github == nil {
			continue
		}

		provider, providerProblems := parseProvider(pr.Github, pr)
		problems = append(problems, providerProblems...)
		if provider != nil {
			providers[*pr.Github] = provider
		}
	}
	return providers, problems
}

func getStateConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     "github-state-cookie",
		Path:     "/",
		MaxAge:   120, // 120 seconds
		HTTPOnly: true,
		Secure:   conf.IsExternalURLSecure(),
	}
	return cfg
}
