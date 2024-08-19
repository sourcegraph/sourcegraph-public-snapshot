package codeintel

import (
	"database/sql"

	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func InitDB(observationCtx *observation.Context) (codeintelshared.CodeIntelDB, error) {
	rawDB, err := initDBMemo.Init(observationCtx)
	if err != nil {
		return nil, err
	}

	return codeintelshared.NewCodeIntelDB(observationCtx.Logger, rawDB), nil
}

var initDBMemo = memo.NewMemoizedConstructorWithArg(func(observationCtx *observation.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	db, err := connections.EnsureNewCodeIntelDB(observationCtx, dsn, "worker")
	if err != nil {
		return nil, errors.Errorf("failed to connect to codeintel database: %s", err)
	}

	return db, nil
})
