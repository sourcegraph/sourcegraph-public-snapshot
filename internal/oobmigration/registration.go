package oobmigration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type RegisterMigratorsFunc func(ctx context.Context, db database.DB, runner *Runner) error

func ComposeRegisterMigratorsFuncs(fns ...RegisterMigratorsFunc) RegisterMigratorsFunc {
	return func(ctx context.Context, db database.DB, runner *Runner) error {
		for _, fn := range fns {
			if fn == nil {
				continue
			}

			if err := fn(ctx, db, runner); err != nil {
				return err
			}
		}

		return nil
	}
}
