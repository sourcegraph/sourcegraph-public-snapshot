package shared

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
)

type registerMigratorsUsingConfAndStoreFactoryFunc func(
	ctx context.Context,
	db database.DB,
	runner *oobmigration.Runner,
	conf conftypes.UnifiedQuerier,
	storeFactory migrations.StoreFactory,
) error

func composeRegisterMigratorsFuncs(fnsFromConfAndStoreFactory ...registerMigratorsUsingConfAndStoreFactoryFunc) func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc {
	return func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc {
		return func(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
			conf, err := newStaticConf(ctx, db)
			if err != nil {
				return err
			}

			fns := make([]oobmigration.RegisterMigratorsFunc, 0, len(fnsFromConfAndStoreFactory))
			for _, f := range fnsFromConfAndStoreFactory {
				f := f // avoid loop capture
				if f == nil {
					continue
				}

				fns = append(fns, func(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
					return f(ctx, db, runner, conf, storeFactory)
				})
			}

			return oobmigration.ComposeRegisterMigratorsFuncs(fns...)(ctx, db, runner)
		}
	}
}
