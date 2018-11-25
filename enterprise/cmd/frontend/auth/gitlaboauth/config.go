package gitlaboauth

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
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

	var (
		mu  sync.Mutex
		cur = map[schema.GitLabAuthProvider]auth.Provider{} // tracks current mapping of valid config to auth.Provider
	)

	go func() {
		conf.Watch(func() {
			mu.Lock()
			defer mu.Unlock()

			if !conf.Get().ExperimentalFeatures.GitlabAuth {
				new := map[schema.GitLabAuthProvider]auth.Provider{}
				updates := make(map[auth.Provider]bool)
				for c, p := range cur {
					if _, ok := new[c]; !ok {
						updates[p] = false
					}
				}
				auth.UpdateProviders(updates)
				cur = new
				ffIsEnabled = false
				return
			}

			log15.Info("Reloading changed GitLab OAuth authentication provider configuration.")

			new, _ := parseConfig(conf.Get())
			updates := make(map[auth.Provider]bool)
			for c, p := range cur {
				if _, ok := new[c]; !ok {
					updates[p] = false
				}
			}
			for c, p := range new {
				if _, ok := cur[c]; !ok {
					updates[p] = true
				}
			}
			auth.UpdateProviders(updates)
			cur = new
			ffIsEnabled = true
		})
	}()
	conf.ContributeValidator(func(cfg schema.SiteConfiguration) (problems []string) {
		_, problems = parseConfig(&cfg)
		return problems
	})
}

func parseConfig(cfg *schema.SiteConfiguration) (providers map[schema.GitLabAuthProvider]auth.Provider, problems []string) {
	providers = make(map[schema.GitLabAuthProvider]auth.Provider)
	for _, pr := range cfg.AuthProviders {
		if pr.Gitlab == nil {
			continue
		}

		if cfg.ExternalURL == "" {
			problems = append(problems, "`externalURL` was empty and it is needed to determine the OAuth callback URL.")
			continue
		}
		externalURL, err := url.Parse(cfg.ExternalURL)
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
