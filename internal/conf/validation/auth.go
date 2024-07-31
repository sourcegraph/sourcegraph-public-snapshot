package validation

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func init() {
	conf.ContributeValidator(validateAuthProvidersConfig)
	conf.ContributeValidator(validateBuiltinAuthConfig)
}

func validateAuthzProviders(cfg conftypes.SiteConfigQuerier, db database.DB) (problems conf.Problems) {
	providers, seriousProblems, warnings, _ := providers.ProvidersFromConfig(context.Background(), cfg, db)
	problems = append(problems, conf.NewExternalServiceProblems(seriousProblems...)...)

	// Validating the connection may make a cross service call, so we should use an
	// internal actor.
	ctx := actor.WithInternalActor(context.Background())

	// Add connection validation issue
	for _, p := range providers {
		if err := p.ValidateConnection(ctx); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s provider %q: %s", p.ServiceType(), p.ServiceID(), err))
		}
	}

	problems = append(problems, conf.NewExternalServiceProblems(warnings...)...)
	return problems
}

func validateAuthProvidersConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if len(c.SiteConfig().AuthProviders) == 0 {
		problems = append(problems, conf.NewSiteProblem("no auth providers set (all access will be forbidden)"))
	}

	// Validate that `auth.enableUsernameChanges` is not set if SSO is configured
	if conf.HasExternalAuthProvider(c) && c.SiteConfig().AuthEnableUsernameChanges {
		problems = append(problems, conf.NewSiteProblem("`auth.enableUsernameChanges` must not be true if external auth providers are set in `auth.providers`"))
	}

	return problems
}

func validateBuiltinAuthConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	var builtinAuthProviders int
	for _, p := range c.SiteConfig().AuthProviders {
		if p.Builtin != nil {
			builtinAuthProviders++
		}
	}
	if builtinAuthProviders >= 2 {
		problems = append(problems, conf.NewSiteProblem(`at most 1 builtin auth provider may be used`))
	}
	return problems
}
