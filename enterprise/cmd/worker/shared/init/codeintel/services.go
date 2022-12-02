package codeintel

import (
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InitServices initializes and returns code intelligence services.
func InitServices(observationContext *observation.Context) (codeintel.Services, error) {
	db, err := workerdb.InitDB(observationContext)
	if err != nil {
		return codeintel.Services{}, err
	}

	codeIntelDB, err := InitDBWithLogger(observationContext.Logger, observationContext)
	if err != nil {
		return codeintel.Services{}, err
	}

	return codeintel.NewServices(codeintel.ServiceDependencies{
		DB:                 db,
		CodeIntelDB:        codeIntelDB,
		ObservationContext: observationContext,
	})
}
