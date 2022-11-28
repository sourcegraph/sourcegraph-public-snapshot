package codeintel

import (
	"database/sql"

	"github.com/sourcegraph/log"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// InitDB initializes and returns a connection to the codeintel db.
func InitDB(observationContext *observation.Context) (*sql.DB, error) {
	return initDBMemo.Init(observationContext)
}

func InitDBWithLogger(logger log.Logger, observationContext *observation.Context) (codeintelshared.CodeIntelDB, error) {
	rawDB, err := InitDB(observationContext)
	if err != nil {
		return nil, err
	}

	return codeintelshared.NewCodeIntelDB(rawDB, logger), nil
}

var initDBMemo = memo.NewMemoizedConstructorWithArg(func(observationContext *observation.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(dsn, "worker", observationContext)
	if err != nil {
		return nil, errors.Errorf("failed to connect to codeintel database: %s", err)
	}

	return db, nil
})
