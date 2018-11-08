package auth

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.UnifiedConfiguration) (problems []string) {
	if len(c.Core.AuthProviders) == 0 {
		problems = append(problems, "no auth providers set (all access will be forbidden)")
	}
	return problems
}
