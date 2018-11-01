package auth

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(config conf.SiteConfiguration) (problems []string) {
	if len(config.Core.AuthProviders) == 0 {
		problems = append(problems, "no auth providers set (all access will be forbidden)")
	}
	return problems
}
