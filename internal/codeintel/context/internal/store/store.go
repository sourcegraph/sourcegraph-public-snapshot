pbckbge store

import (
	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	// TODO
}

type store struct {
	db         *bbsestore.Store
	logger     logger.Logger
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("context.store", ""),
		operbtions: newOperbtions(observbtionCtx),
	}
}
