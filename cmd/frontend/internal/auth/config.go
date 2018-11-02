package auth

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conftypes.SiteConfiguration) (problems []string) {
	if len(c.AuthProviders) == 0 {
		problems = append(problems, "no auth providers set (all access will be forbidden)")
	}
	return problems
}
