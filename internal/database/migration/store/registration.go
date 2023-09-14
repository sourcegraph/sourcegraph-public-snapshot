package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type RegisterMigratorsUsingConfAndStoreFactoryFunc func(
	ctx context.Context,
	db database.DB,
	runner *oobmigration.Runner,
	conf conftypes.UnifiedQuerier,
	storeFactory migrations.StoreFactory,
) error

func ComposeRegisterMigratorsFuncs(fnsFromConfAndStoreFactory ...RegisterMigratorsUsingConfAndStoreFactoryFunc) func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc {
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

type staticConf struct {
	serviceConnections conftypes.ServiceConnections
	siteConfig         schema.SiteConfiguration
}

func newStaticConf(ctx context.Context, db database.DB) (*staticConf, error) {
	serviceConnections, err := migrationServiceConnections()
	if err != nil {
		return nil, err
	}

	siteConfig, err := siteConfig(ctx, basestore.NewWithHandle(db.Handle()))
	if err != nil {
		return nil, err
	}

	return &staticConf{
		serviceConnections: serviceConnections,
		siteConfig:         siteConfig,
	}, nil
}

func (c staticConf) ServiceConnections() conftypes.ServiceConnections { return c.serviceConnections }
func (c staticConf) SiteConfig() schema.SiteConfiguration             { return c.siteConfig }

func migrationServiceConnections() (conftypes.ServiceConnections, error) {
	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		return conftypes.ServiceConnections{}, err
	}

	return conftypes.ServiceConnections{
		PostgresDSN:          dsns["frontend"],
		CodeIntelPostgresDSN: dsns["codeintel"],
		CodeInsightsDSN:      dsns["codeinsights"],
	}, nil
}

func siteConfig(ctx context.Context, store *basestore.Store) (siteConfig schema.SiteConfiguration, _ error) {
	raw, ok, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(siteConfigQuery)))
	if err != nil {
		return siteConfig, err
	}
	if !ok {
		return siteConfig, errors.New("instance is new")
	}

	if err := jsonc.Unmarshal(raw, &siteConfig); err != nil {
		return siteConfig, err
	}

	return siteConfig, nil
}

const siteConfigQuery = `
SELECT c.contents
FROM critical_and_site_config c
WHERE c.type = 'site'
ORDER BY c.id DESC
LIMIT 1
`
