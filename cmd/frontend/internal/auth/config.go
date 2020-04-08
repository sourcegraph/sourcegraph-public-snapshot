package auth

import "github.com/sourcegraph/sourcegraph/internal/conf"

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.Unified) (problems conf.Problems) {
	if len(c.AuthProviders) == 0 {
		problems = append(problems, conf.NewSiteProblem("no auth providers set (all access will be forbidden)"))
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.AuthEnableUsernameChanges {
		problems = append(problems, conf.NewSiteProblem("`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`"))
	}

	return problems
}
