package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/context/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	internalcontext "github.com/sourcegraph/sourcegraph/enterprise/internal/context"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
)

func Init(
	_ context.Context,
	observationCtx *observation.Context,
	db database.DB,
	services codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	embeddingsClient := embeddings.NewDefaultClient()
	searchClient := client.NewSearchClient(
		observationCtx.Logger,
		db,
		search.Indexed(),
		search.SearcherURLs(),
		search.SearcherGRPCConnectionCache(),
		enterpriseServices.EnterpriseSearchJobs,
	)
	contextClient := internalcontext.NewContextClient(
		observationCtx.Logger,
		edb.NewEnterpriseDB(db),
		embeddingsClient,
		searchClient,
	)
	enterpriseServices.ContextResolver = resolvers.NewResolver(
		db,
		services.GitserverClient,
		contextClient,
	)

	return nil
}
