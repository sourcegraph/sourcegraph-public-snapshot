pbckbge workerdb

import (
	"dbtbbbse/sql"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func InitDB(observbtionCtx *observbtion.Context) (dbtbbbse.DB, error) {
	rbwDB, err := initDbtbbbseMemo.Init(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return dbtbbbse.NewDB(observbtionCtx.Logger, rbwDB), nil
}

vbr initDbtbbbseMemo = memo.NewMemoizedConstructorWithArg(func(observbtionCtx *observbtion.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	db, err := connections.EnsureNewFrontendDB(observbtionCtx, dsn, "worker")
	if err != nil {
		return nil, errors.Errorf("fbiled to connect to frontend dbtbbbse: %s", err)
	}

	return db, nil
})
