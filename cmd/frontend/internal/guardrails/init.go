package guardrails

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	resolver := &resolvers.GuardrailsResolver{}
	if envvar.SourcegraphDotComMode() {
		// On DotCom guardrails endpoint runs search.
		resolver.AttributionService = initDotcomAttributionService(observationCtx, db)
	} else {
		// On an Enterprise instance endpoint proxies to gateway.
		resolver.AttributionService = initEnterpriseAttributionService(observationCtx)
	}
	enterpriseServices.GuardrailsResolver = resolver
	return nil
}

func initEnterpriseAttributionService(observationCtx *observation.Context) attribution.Service {
	client, ok := codygateway.NewClientFromSiteConfig(httpcli.ExternalDoer)
	if !ok {
		// TODO(#59701) handle error
		return nil
	}
	return attribution.NewGatewayProxy(observationCtx, client)
}

func initDotcomAttributionService(observationCtx *observation.Context, db database.DB) attribution.Service {
	searchClient := client.New(observationCtx.Logger, db, gitserver.NewClient("http.guardrails.search"))
	return attribution.NewLocalSearch(observationCtx, searchClient)
}

func alwaysAllowed(context.Context, string) (bool, error) {
	return true, nil
}

func NewAttributionTest(observationCtx *observation.Context) func(context.Context, string) (bool, error) {
	// TODO(#59701): Re-initialize attribution service. So that changes
	// in site-config are reflected immediately for subsequent GraphQL
	// calls and code completions calls.
	service := initEnterpriseAttributionService(observationCtx)
	if service == nil {
		return alwaysAllowed
	}
	// Attribution is only-enterprise, dotcom lets everything through.
	if envvar.SourcegraphDotComMode() {
		return alwaysAllowed
	}
	return func(ctx context.Context, snippet string) (bool, error) {
		// Check if attribution is on, permit everything if it's off.
		c := conf.GetConfigFeatures(conf.Get().SiteConfig())
		if !c.Attribution {
			return true, nil
		}
		attribution, err := service.SnippetAttribution(ctx, snippet, 1)
		// Attribution not available. Mode is permissive.
		if err != nil {
			return true, err
		}
		// Permit completion if no attribution found.
		return len(attribution.RepositoryNames) == 0, nil
	}
}
