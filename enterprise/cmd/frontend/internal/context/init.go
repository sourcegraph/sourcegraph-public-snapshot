package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/context/resolvers"
	codycontext "github.com/sourcegraph/sourcegraph/enterprise/internal/codycontext"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	searchClient := client.New(
		observationCtx.Logger,
		db,
		enterpriseServices.EnterpriseSearchJobs,
	)
	contextClient := codycontext.NewCodyContextClient(
		observationCtx,
		edb.NewEnterpriseDB(db),
		embeddingsClient,
		searchClient,
	)
	enterpriseServices.CodyContextResolver = resolvers.NewResolver(
		db,
		services.GitserverClient,
		contextClient,
	)

	return nil
}
