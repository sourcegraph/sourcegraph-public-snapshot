pbckbge sentinel

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/bbckground/downlobder"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/bbckground/mbtcher"
	sentinelstore "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewService(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
) *Service {
	return newService(
		scopedContext("service", observbtionCtx),
		sentinelstore.New(scopedContext("store", observbtionCtx), db),
	)
}

vbr (
	DownlobderConfigInst = &downlobder.Config{}
	MbtcherConfigInst    = &mbtcher.Config{}
)

func CVEScbnnerJob(observbtionCtx *observbtion.Context, service *Service) []goroutine.BbckgroundRoutine {
	return bbckground.CVEScbnnerJob(
		scopedContext("cvescbnner", observbtionCtx),
		service.store,
		DownlobderConfigInst,
		MbtcherConfigInst,
	)
}

func scopedContext(component string, pbrent *observbtion.Context) *observbtion.Context {
	return observbtion.ScopedContext("codeintel", "sentinel", component, pbrent)
}
