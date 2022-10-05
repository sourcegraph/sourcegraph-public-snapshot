package codeintel

import (
	"github.com/sourcegraph/log"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
)

// InitServices initializes and returns code intelligence services.
func InitServices() (*codeintel.Services, error) {
	databases, err := databases(log.Scoped("codeintel worker", ""))
	if err != nil {
		return nil, err
	}

	return codeintel.GetServices(databases)
}

func databases(logger log.Logger) (codeintel.Databases, error) {
	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return codeintel.Databases{}, err
	}

	codeIntelDB, err := InitDBWithLogger(logger)
	if err != nil {
		return codeintel.Databases{}, err
	}

	return codeintel.Databases{
		DB:          db,
		CodeIntelDB: codeIntelDB,
	}, nil
}
