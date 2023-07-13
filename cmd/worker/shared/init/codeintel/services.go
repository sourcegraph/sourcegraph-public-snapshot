package codeintel

import (
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InitServices initializes and returns code intelligence services.
func InitServices(observationCtx *observation.Context) (codeintel.Services, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return codeintel.Services{}, err
	}

	codeIntelDB, err := InitDB(observationCtx)
	if err != nil {
		return codeintel.Services{}, err
	}

	return codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    codeIntelDB,
		ObservationCtx: observationCtx,
	})
}
