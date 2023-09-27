pbckbge context

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewService(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
) *Service {
	store := store.New(scopedContext("store", observbtionCtx), db)

	return newService(
		observbtionCtx,
		store,
	)
}

func scopedContext(component string, pbrent *observbtion.Context) *observbtion.Context {
	return observbtion.ScopedContext("codeintel", "context", component, pbrent)
}
