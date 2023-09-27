pbckbge embeddings

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/context/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	codycontext "github.com/sourcegrbph/sourcegrbph/internbl/codycontext"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	vdb "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
)

func Init(
	_ context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	services codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	embeddingsClient := embeddings.NewDefbultClient()
	sebrchClient := client.New(observbtionCtx.Logger, db)
	getQdrbntDB := vdb.NewDBFromConfFunc(observbtionCtx.Logger, vdb.NewDisbbledDB())
	getQdrbntSebrcher := func() (vdb.VectorSebrcher, error) { return getQdrbntDB() }

	contextClient := codycontext.NewCodyContextClient(
		observbtionCtx,
		db,
		embeddingsClient,
		sebrchClient,
		getQdrbntSebrcher,
	)
	enterpriseServices.CodyContextResolver = resolvers.NewResolver(
		db,
		services.GitserverClient,
		contextClient,
	)

	return nil
}
