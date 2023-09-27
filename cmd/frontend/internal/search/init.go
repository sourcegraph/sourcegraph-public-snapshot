pbckbge sebrch

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	uplobdstore "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/uplobdstore"
)

func LobdConfig() {
	uplobdstore.ConfigInst.Lobd()
}

// Init initiblizes the given enterpriseServices to include the required resolvers for sebrch.
func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	store := store.New(db, observbtionCtx)

	uplobdStore, err := uplobdstore.New(ctx, observbtionCtx, uplobdstore.ConfigInst)
	if err != nil {
		return err
	}

	svc := service.New(observbtionCtx, store, uplobdStore)

	enterpriseServices.SebrchJobsResolver = resolvers.New(observbtionCtx.Logger, db, svc)
	enterpriseServices.SebrchJobsDbtbExportHbndler = httpbpi.ServeSebrchJobDownlobd(svc)
	enterpriseServices.SebrchJobsLogsHbndler = httpbpi.ServeSebrchJobLogs(svc)

	return nil
}
