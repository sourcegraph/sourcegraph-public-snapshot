package shared

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	conf.ContributeValidator(func(cfg schema.SiteConfiguration) []string {
		_, _, seriousProblems, warnings := providersFromConfig(&cfg)
		return append(seriousProblems, warnings...)
	})
	go conf.Watch(func() {
		allowAccessByDefault, authzProviders, _, _ := providersFromConfig(conf.Get())
		authz.SetProviders(allowAccessByDefault, authzProviders)
	})
}

// providersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings".  "Serious problems" are those that should make Sourcegraph set
// authz.allowAccessByDefault to false. "Warnings" are all other validation problems.
func providersFromConfig(cfg *schema.SiteConfiguration) (
	allowAccessByDefault bool,
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	allowAccessByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.")
			allowAccessByDefault = false
		}
	}()

	glp, glproblems, glwarnings := gitlabProvidersFromConfig(cfg)
	authzProviders = append(authzProviders, glp...)
	seriousProblems = append(seriousProblems, glproblems...)
	warnings = append(warnings, glwarnings...)

	ghp, ghproblems, ghwarnings := githubProvidersFromConfig(cfg)
	authzProviders = append(authzProviders, ghp...)
	seriousProblems = append(seriousProblems, ghproblems...)
	warnings = append(warnings, ghwarnings...)

	return allowAccessByDefault, authzProviders, seriousProblems, warnings
}
