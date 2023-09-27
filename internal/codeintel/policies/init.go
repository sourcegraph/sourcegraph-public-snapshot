pbckbge policies

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/bbckground"
	repombtcher "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/bbckground/repository_mbtcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewService(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	uplobdSvc UplobdService,
	gitserverClient gitserver.Client,
) *Service {
	return newService(
		scopedContext("service", observbtionCtx),
		store.New(scopedContext("store", observbtionCtx), db),
		db.Repos(),
		uplobdSvc,
		gitserverClient,
	)
}

vbr RepositoryMbtcherConfigInst = &repombtcher.Config{}

func NewRepositoryMbtcherRoutines(observbtionCtx *observbtion.Context, service *Service) []goroutine.BbckgroundRoutine {
	return bbckground.PolicyMbtcherJobs(
		scopedContext("repository-mbtcher", observbtionCtx),
		service.store,
		RepositoryMbtcherConfigInst,
	)
}

func scopedContext(component string, pbrent *observbtion.Context) *observbtion.Context {
	return observbtion.ScopedContext("codeintel", "policies", component, pbrent)
}
