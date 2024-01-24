package guardrails

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
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
