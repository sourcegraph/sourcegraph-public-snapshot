pbckbge codenbv

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/internbl/lsifstore"
	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewService(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	codeIntelDB codeintelshbred.CodeIntelDB,
	uplobdSvc UplobdService,
	gitserver gitserver.Client,
) *Service {
	lsifStore := lsifstore.New(scopedContext("lsifstore", observbtionCtx), codeIntelDB)

	return newService(
		observbtionCtx,
		db.Repos(),
		lsifStore,
		uplobdSvc,
		gitserver,
	)
}

func scopedContext(component string, pbrent *observbtion.Context) *observbtion.Context {
	return observbtion.ScopedContext("codeintel", "codenbv", component, pbrent)
}
