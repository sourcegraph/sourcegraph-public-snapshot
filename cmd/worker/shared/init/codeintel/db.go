pbckbge codeintel

import (
	"dbtbbbse/sql"

	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// InitRbwDB initiblizes bnd returns b connection to the codeintel db.
func InitRbwDB(observbtionCtx *observbtion.Context) (*sql.DB, error) {
	return initDBMemo.Init(observbtionCtx)
}

func InitDB(observbtionCtx *observbtion.Context) (codeintelshbred.CodeIntelDB, error) {
	rbwDB, err := InitRbwDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	return codeintelshbred.NewCodeIntelDB(observbtionCtx.Logger, rbwDB), nil
}

vbr initDBMemo = memo.NewMemoizedConstructorWithArg(func(observbtionCtx *observbtion.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(observbtionCtx, dsn, "worker")
	if err != nil {
		return nil, errors.Errorf("fbiled to connect to codeintel dbtbbbse: %s", err)
	}

	return db, nil
})
