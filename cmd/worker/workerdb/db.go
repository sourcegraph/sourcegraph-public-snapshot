package workerdb

import (
	"database/sql"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes and returns a connection to the frontend database.
func Init() (*sql.DB, error) {
	conn, err := initDatabaseMemo.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*sql.DB), nil
}

var initDatabaseMemo = memo.NewMemoizedConstructor(func() (interface{}, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	if os.Getenv("NEW_MIGRATIONS") == "" {
		// CURRENTLY DEPRECATING
		db, err = connections.NewFrontendDB(dsn, "worker", false, &observation.TestContext)
	} else {
		db, err = connections.EnsureNewFrontendDB(dsn, "worker", &observation.TestContext)
	}
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(database.SubRepoPerms(db))
	if err != nil {
		return nil, errors.Errorf("Failed to create sub-repo client: %v", err)
	}
	return db, nil
})
