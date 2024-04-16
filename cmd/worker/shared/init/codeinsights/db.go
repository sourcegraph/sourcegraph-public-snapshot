package codeintel

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

// InitRawDB initializes and returns a connection to the codeinsights db.
func InitRawDB(observationCtx *observation.Context) (*sql.DB, error) {
	return initDBMemo.Init(observationCtx)
}

func InitDB(observationCtx *observation.Context) (database.InsightsDB, error) {
	rawDB, err := InitRawDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return database.NewInsightsDB(rawDB, observationCtx.Logger), nil
}

var initDBMemo = memo.NewMemoizedConstructorWithArg(func(observationCtx *observation.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeInsightsDSN
	})

	db, err := connections.EnsureNewCodeInsightsDB(observationCtx, dsn, "worker")
	if err != nil {
		return nil, errors.Errorf("failed to connect to codeinsights database: %s", err)
	}

	return db, nil
})
