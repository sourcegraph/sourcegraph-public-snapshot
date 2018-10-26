package auth

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c schema.SiteConfiguration) (problems []string) {
	if c.AuthProvider != "" && len(c.AuthProviders) > 0 {
		problems = append(problems, `auth.providers takes precedence over auth.provider (deprecated) when both are set (auth.provider is IGNORED in that case)`)
	} else if c.AuthProvider != "" {
		problems = append(problems, `auth.provider is deprecated; use auth.providers instead`)
	}

	authProviders := conf.AuthProvidersFromConfig(&c)
	if len(authProviders) == 0 {
		problems = append(problems, "no auth providers set (all access will be forbidden)")
	}

	return problems
}
