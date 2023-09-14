package auth

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// TODO: explicit validateconf.Init call instead of implicit
func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if len(c.SiteConfig().AuthProviders) == 0 {
		problems = append(problems, conf.NewSiteProblem("no auth providers set (all access will be forbidden)"))
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.SiteConfig().AuthEnableUsernameChanges {
		problems = append(problems, conf.NewSiteProblem("`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`"))
	}

	return problems
}
