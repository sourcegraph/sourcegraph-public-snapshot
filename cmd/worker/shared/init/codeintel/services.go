pbckbge codeintel

import (
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// InitServices initiblizes bnd returns code intelligence services.
func InitServices(observbtionCtx *observbtion.Context) (codeintel.Services, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return codeintel.Services{}, err
	}

	codeIntelDB, err := InitDB(observbtionCtx)
	if err != nil {
		return codeintel.Services{}, err
	}

	return codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    codeIntelDB,
		ObservbtionCtx: observbtionCtx,
	})
}
