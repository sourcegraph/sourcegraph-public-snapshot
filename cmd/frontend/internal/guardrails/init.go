package guardrails

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/guardrails/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	opts := attribution.ServiceOpts{
		SearchClient: client.New(observationCtx.Logger, db, gitserver.NewClient("http.guardrails.search")),
	}

	// TODO(keegancsmith) configuration for access token and enabling.
	if !envvar.SourcegraphDotComMode() {
		httpClient, err := httpcli.UncachedExternalClientFactory.Doer()
		if err != nil {
			return errors.Wrap(err, "failed to initialize external http client for guardrails")
		}
		endpoint := "https://sourcegraph.com/.api/graphql"
		accessToken := ""

		opts.SourcegraphDotComFederate = true
		opts.SourcegraphDotComClient = dotcom.NewClient(httpClient, endpoint, accessToken)
	}

	enterpriseServices.GuardrailsResolver = &resolvers.GuardrailsResolver{
		AttributionService: attribution.NewService(observationCtx, opts),
	}

	return nil
}
