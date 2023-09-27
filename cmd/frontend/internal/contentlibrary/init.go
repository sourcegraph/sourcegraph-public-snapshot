pbckbge contentlibrbry

import (
	"context"

	logger "github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// Init initiblizes the given enterpriseServices to include the required
// resolvers for the sebrch content librbry.
func Init(
	_ context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	logger := logger.Scoped("contentlibrbry", "sourcegrbph content librbry")
	enterpriseServices.ContentLibrbryResolver = grbphqlbbckend.NewContentLibrbryResolver(db, logger)
	return nil
}
