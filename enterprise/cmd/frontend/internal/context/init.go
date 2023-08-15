package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/context/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	codycontext "github.com/sourcegraph/sourcegraph/internal/codycontext"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	vdb "github.com/sourcegraph/sourcegraph/internal/embeddings/db"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
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
	searchClient := client.New(observationCtx.Logger, db)
	qdrantSearcher := vdb.NewDisabledDB()
	if addr := conf.ServiceConnections().Qdrant; addr != "" {
		conn, err := defaults.Dial(addr, observationCtx.Logger)
		if err != nil {
			return err
		}
		qdrantSearcher = vdb.NewQdrantDBFromConn(conn)
	}
	contextClient := codycontext.NewCodyContextClient(
		observationCtx,
		db,
		embeddingsClient,
		searchClient,
		qdrantSearcher,
	)
	enterpriseServices.CodyContextResolver = resolvers.NewResolver(
		db,
		services.GitserverClient,
		contextClient,
	)

	return nil
}
