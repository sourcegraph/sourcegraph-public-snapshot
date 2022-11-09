package codeintel

import (
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
)

// InitServices initializes and returns code intelligence services.
func InitServices() (codeintel.Services, error) {
	logger := log.Scoped("codeintel", "codeintel services")

	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return codeintel.Services{}, err
	}

	codeIntelDB, err := InitDBWithLogger(logger)
	if err != nil {
		return codeintel.Services{}, err
	}

	return codeintel.GetServices(codeintel.Databases{
		DB:          db,
		CodeIntelDB: codeIntelDB,
	})
}
