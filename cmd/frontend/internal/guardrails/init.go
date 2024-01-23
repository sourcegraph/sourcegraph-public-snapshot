package guardrails

import (
	"context"
	"fmt"
	"strings"
	"time"

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

type ZeroAttributionFilter struct {
	service attribution.Service
}

func NewAttributionFilter(observationCtx *observation.Context) ZeroAttributionFilter {
	service := initEnterpriseAttributionService(observationCtx)
	return ZeroAttributionFilter{service: service}
}

func (f ZeroAttributionFilter) InScope(snippet string) bool {
	lines := strings.Split(snippet, "\n")
	return len(lines) > 10
}

func (f ZeroAttributionFilter) CanUse(ctx context.Context, snippet string, limit int) bool {
	// TODO
	start := time.Now()
	if f.service == nil {
		fmt.Println("ATTRIBUTION ERROR")
		return true
	}
	// Attribution is only-enterprise, dotcom lets everything through.
	if envvar.SourcegraphDotComMode() {
		return true
	}
	// Check if attribution is on, permit everything if it's off.
	c := conf.GetConfigFeatures(conf.Get().SiteConfig())
	if !c.Attribution {
		return true
	}
	// if len(snippet) > 50 {
	// 	snippet = snippet[:50]
	// }
	defer func () {
		now := time.Now()
		fmt.Printf("ATTRIBUTION TIME %d characters %s\n", len(snippet), now.Sub(start))
		fmt.Println(snippet)
		}()
	attribution, err := f.service.SnippetAttribution(ctx, snippet, 1)
	// Attribution not available. Mode is permissive.
	if err != nil {
		return true
	}
	// Permit completion if no attribution found.
	return len(attribution.RepositoryNames) == 0
}
