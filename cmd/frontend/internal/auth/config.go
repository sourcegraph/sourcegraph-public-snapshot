package auth

import "github.com/sourcegraph/sourcegraph/pkg/conf"

func init() {
	conf.ContributeValidator(validateConfig)
}

func validateConfig(c conf.Unified) (problems conf.Problems) {
	if len(c.Critical.AuthProviders) == 0 {
		problems = append(problems, conf.NewCriticalProblem("no auth providers set (all access will be forbidden)"))
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.Critical.AuthEnableUsernameChanges {
		problems = append(problems, conf.NewCriticalProblem("`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`"))
	}

	return problems
}
