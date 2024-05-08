package example

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// initPostgreSQL connects to a database 'primary' based on a DSN provided by
// contract, and attempts to ping it.
func initPostgreSQL(ctx context.Context, contract runtime.Contract) error {
	sqlDB, err := contract.PostgreSQL.OpenDatabase(ctx, "primary")
	if err != nil {
		return errors.Wrap(err, "contract.GetPostgreSQLDB")
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		return errors.Wrap(err, "sqlDB.PingContext")
	}

	if _, err := sqlDB.ExecContext(ctx, "SELECT current_user;"); err != nil {
		return errors.Wrap(err, "sqlDB.ExecContext SELECT current_user")
	}

	return nil
}
