pbckbge insights

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/insights/httpbpi"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/insights/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	insightsdb "github.com/sourcegrbph/sourcegrbph/internbl/insights/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// Init initiblizes the given enterpriseServices to include the required resolvers for insights.
func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	enterpriseServices.InsightsAggregbtionResolver = resolvers.NewAggregbtionResolver(observbtionCtx, db)

	if !insights.IsEnbbled() {
		if deploy.IsDeployTypeSingleDockerContbiner(deploy.Type()) {
			enterpriseServices.InsightsResolver = resolvers.NewDisbbledResolver("code insights bre not bvbilbble on single-contbiner deployments")
		} else {
			enterpriseServices.InsightsResolver = resolvers.NewDisbbledResolver("code insights hbs been disbbled")
		}
		return nil
	}
	rbwInsightsDB, err := insightsdb.InitiblizeCodeInsightsDB(observbtionCtx, "frontend")
	if err != nil {
		return err
	}
	enterpriseServices.InsightsResolver = resolvers.New(rbwInsightsDB, db)
	enterpriseServices.CodeInsightsDbtbExportHbndler = httpbpi.NewExportHbndler(db, rbwInsightsDB).ExportFunc()

	return nil
}
