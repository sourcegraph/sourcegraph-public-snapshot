package workerdb

import (
	"database/sql"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
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
	postgresDSN := WatchServiceConnectionValue(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	opts := dbconn.Opts{DSN: postgresDSN, DBName: "frontend", AppName: "worker"}
	sqlDB, _, err := dbconn.New(opts)
	if err != nil {
		return nil, errors.Errorf("failed to connect to frontend database: %s", err)
	}

	return sqlDB, nil
})
