package guardrails

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/guardrails/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/settings"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	attributionService := &attribution.Service{
		CurrentUserFinal:      settingsService(db),
		SearchClient:          client.New(observationCtx.Logger, db, enterpriseServices.EnterpriseSearchJobs),
		SourcegraphDotComMode: envvar.SourcegraphDotComMode(),
	}

	enterpriseServices.GuardrailsResolver = &resolvers.GuardrailsResolver{
		AttributionService: attributionService,
	}

	return nil
}

// settingsService is a temporary helper until we introduce a settings service.
func settingsService(db database.DB) func(context.Context) (*schema.Settings, error) {
	return func(ctx context.Context) (*schema.Settings, error) {
		return settings.CurrentUserFinal(ctx, db)
	}
}
