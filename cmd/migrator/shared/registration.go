package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type registerMigratorsFromConfFunc func(
	ctx context.Context,
	db database.DB,
	runner *oobmigration.Runner,
	conf conftypes.UnifiedQuerier,
) error

func composeRegisterMigratorsFuncs(fnsFromConfig ...registerMigratorsFromConfFunc) oobmigration.RegisterMigratorsFunc {
	return func(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
		conf, err := newStaticConf(ctx, db)
		if err != nil {
			return err
		}

		fns := make([]oobmigration.RegisterMigratorsFunc, 0, len(fnsFromConfig))
		for _, f := range fnsFromConfig {
			f := f // avoid loop capture

			fns = append(fns, func(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
				return f(ctx, db, runner, conf)
			})
		}

		return oobmigration.ComposeRegisterMigratorsFuncs(fns...)(ctx, db, runner)
	}
}
