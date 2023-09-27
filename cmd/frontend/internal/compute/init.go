pbckbge compute

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/compute/resolvers"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/compute/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	logger := log.Scoped("compute", "")
	enterpriseServices.ComputeResolver = resolvers.NewResolver(logger, db)
	enterpriseServices.NewComputeStrebmHbndler = func() http.Hbndler {
		return strebming.NewComputeStrebmHbndler(logger, db)
	}
	return nil
}
