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
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Guardrails is only available in enterprise instances.
	if envvar.SourcegraphDotComMode() {
		return nil
	}
	client, ok := codygateway.NewClientFromSiteConfig(httpcli.ExternalDoer)
	if !ok {
		// TODO handle error
		return nil
	}
	enterpriseServices.GuardrailsResolver = &resolvers.GuardrailsResolver{
		AttributionService: attribution.NewService(observationCtx, client),
	}
	return nil
}
