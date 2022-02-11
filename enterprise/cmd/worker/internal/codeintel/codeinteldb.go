package codeintel

import (
	"database/sql"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/worker/memo"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// InitCodeIntelDatabase initializes and returns a connection to the codeintel db.
func InitCodeIntelDatabase() (*sql.DB, error) {
	conn, err := initCodeIntelDatabaseMemo.Init()
	if err != nil {
		return nil, err
	}

	return conn.(*sql.DB), err
}

var initCodeIntelDatabaseMemo = memo.NewMemoizedConstructor(func() (interface{}, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	if os.Getenv("NEW_MIGRATIONS") == "" {
		// CURRENTLY DEPRECATING
		db, err = connections.NewCodeIntelDB(dsn, "worker", false, &observation.TestContext)
	} else {
		db, err = connections.EnsureNewCodeIntelDB(dsn, "worker", &observation.TestContext)
	}
	if err != nil {
		return nil, errors.Errorf("failed to connect to codeintel database: %s", err)
	}

	return db, nil
})
