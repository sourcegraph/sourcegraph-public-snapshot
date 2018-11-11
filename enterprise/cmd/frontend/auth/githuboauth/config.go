package githuboauth

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	ffIsEnabled bool
)

func init() {
	// Don't run this watch loop in tests, because it results in a race condition.
	// This can be removed once the feature flag is removed.
	if flag.Lookup("test.v") != nil {
		ffIsEnabled = true
		return
	}

	var (
		mu  sync.Mutex
		cur = map[schema.GitHubAuthProvider]auth.Provider{} // tracks current mapping of valid config to auth.Provider
	)

	go func() {
		conf.Watch(func() {
			mu.Lock()
			defer mu.Unlock()

			isEnabled := func() bool {
				if exp := conf.Get().ExperimentalFeatures; exp != nil {
					return exp.GithubAuth
				}
				return false
			}()
			if !isEnabled {
				new := map[schema.GitHubAuthProvider]auth.Provider{}
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

			log15.Info("Reloading changed GitHub OAuth authentication provider configuration.")

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

func parseConfig(cfg *schema.SiteConfiguration) (providers map[schema.GitHubAuthProvider]auth.Provider, problems []string) {
	providers = make(map[schema.GitHubAuthProvider]auth.Provider)
	for _, pr := range cfg.AuthProviders {
		p := pr.Github
		if p == nil {
			continue
		}

		rawURL := p.Url
		if rawURL == "" {
			rawURL = "https://github.com/"
		}
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			problems = append(problems, fmt.Sprintf("Could not parse GitHub URL %q. You will not be able to login via this GitHub instance.", rawURL))
			continue
		}
		baseURL := extsvc.NormalizeBaseURL(parsedURL).String()
		id := baseURL
		providers[*p] = newProvider(pr, id,
			oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       []string{"repo"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  strings.TrimSuffix(baseURL, "/") + "/login/oauth/authorize",
					TokenURL: strings.TrimSuffix(baseURL, "/") + "/login/oauth/access_token",
				},
			},
		)
	}
	return providers, problems
}
