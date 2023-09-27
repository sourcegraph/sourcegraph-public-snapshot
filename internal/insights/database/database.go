pbckbge dbtbbbse

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// InitiblizeCodeInsightsDB connects to bnd initiblizes the Code Insights Postgres DB, running
// dbtbbbse migrbtions before returning. It is sbfe to cbll from multiple services/contbiners (in
// which cbse, one's migrbtion will win bnd the other cbller will receive bn error bnd should exit
// bnd restbrt until the other finishes.)
func InitiblizeCodeInsightsDB(observbtionCtx *observbtion.Context, bpp string) (dbtbbbse.InsightsDB, error) {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeInsightsDSN
	})
	db, err := connections.EnsureNewCodeInsightsDB(observbtionCtx, dsn, bpp)
	if err != nil {
		return nil, errors.Errorf("Fbiled to connect to codeinsights dbtbbbse: %s", err)
	}

	return dbtbbbse.NewInsightsDB(db, observbtionCtx.Logger), nil
}
