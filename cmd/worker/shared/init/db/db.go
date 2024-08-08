package workerdb

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func InitDB(observationCtx *observation.Context) (database.DB, error) {
	rawDB, err := initDatabaseMemo.Init(observationCtx)
	if err != nil {
		return nil, err
	}

	return database.NewDB(observationCtx.Logger, rawDB), nil
}

var initDatabaseMemo = memo.NewMemoizedConstructorWithArg(func(observationCtx *observation.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	db, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "worker")
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	return db, nil
})
