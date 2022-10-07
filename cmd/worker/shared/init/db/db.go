package workerdb

import (
	"database/sql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Init initializes and returns a connection to the frontend database.
func Init(observationContext *observation.Context) (*sql.DB, error) {
	return initDatabaseMemo.Init(observationContext)
}

func InitDBWithLogger(logger log.Logger, observationContext *observation.Context) (database.DB, error) {
	rawDB, err := Init(observationContext)
	if err != nil {
		return nil, err
	}

	return database.NewDB(logger, rawDB), nil
}

var initDatabaseMemo = memo.NewMemoizedConstructorWithArg(func(observationContext *observation.Context) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	db, err := connections.EnsureNewFrontendDB(dsn, "worker", observationContext)
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(database.NewDB(log.Scoped("initDatabaseMemo", ""), db).SubRepoPerms())
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}
	return db, nil
})
