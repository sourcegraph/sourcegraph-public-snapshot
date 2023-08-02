package confvalidators

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// TODO(nsc): use c
// Report any authz provider problems in external configs.
func validateAuthzProviders(cfg conftypes.SiteConfigQuerier) (problems conf.Problems) {
	_, providers, seriousProblems, warnings, _ := providers.ProvidersFromConfig(ctx, cfg, extsvcStore, db)
	problems = append(problems, conf.NewExternalServiceProblems(seriousProblems...)...)

	// Validating the connection may make a cross service call, so we should use an
	// internal actor.
	ctx := actor.WithInternalActor(ctx)

	// Add connection validation issue
	for _, p := range providers {
		if err := p.ValidateConnection(ctx); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s provider %q: %s", p.ServiceType(), p.ServiceID(), err))
		}
	}

	problems = append(problems, conf.NewExternalServiceProblems(warnings...)...)
	return problems
}
