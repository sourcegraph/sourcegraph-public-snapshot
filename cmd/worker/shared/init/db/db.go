package workerdb

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Init initializes and returns a connection to the frontend database.
func Init() (*sql.DB, error) {
	return initDatabaseMemo.Init()
}

var initDatabaseMemo = memo.NewMemoizedConstructor(func() (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	db, err := connections.EnsureNewFrontendDB(dsn, "worker", &observation.TestContext)
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(database.SubRepoPerms(db))
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}
	return db, nil
})
