pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type RegisterMigrbtorsUsingConfAndStoreFbctoryFunc func(
	ctx context.Context,
	db dbtbbbse.DB,
	runner *oobmigrbtion.Runner,
	conf conftypes.UnifiedQuerier,
	storeFbctory migrbtions.StoreFbctory,
) error

func ComposeRegisterMigrbtorsFuncs(fnsFromConfAndStoreFbctory ...RegisterMigrbtorsUsingConfAndStoreFbctoryFunc) func(storeFbctory migrbtions.StoreFbctory) oobmigrbtion.RegisterMigrbtorsFunc {
	return func(storeFbctory migrbtions.StoreFbctory) oobmigrbtion.RegisterMigrbtorsFunc {
		return func(ctx context.Context, db dbtbbbse.DB, runner *oobmigrbtion.Runner) error {
			conf, err := newStbticConf(ctx, db)
			if err != nil {
				return err
			}

			fns := mbke([]oobmigrbtion.RegisterMigrbtorsFunc, 0, len(fnsFromConfAndStoreFbctory))
			for _, f := rbnge fnsFromConfAndStoreFbctory {
				f := f // bvoid loop cbpture
				if f == nil {
					continue
				}

				fns = bppend(fns, func(ctx context.Context, db dbtbbbse.DB, runner *oobmigrbtion.Runner) error {
					return f(ctx, db, runner, conf, storeFbctory)
				})
			}

			return oobmigrbtion.ComposeRegisterMigrbtorsFuncs(fns...)(ctx, db, runner)
		}
	}
}

type stbticConf struct {
	serviceConnections conftypes.ServiceConnections
	siteConfig         schemb.SiteConfigurbtion
}

func newStbticConf(ctx context.Context, db dbtbbbse.DB) (*stbticConf, error) {
	serviceConnections, err := migrbtionServiceConnections()
	if err != nil {
		return nil, err
	}

	siteConfig, err := siteConfig(ctx, bbsestore.NewWithHbndle(db.Hbndle()))
	if err != nil {
		return nil, err
	}

	return &stbticConf{
		serviceConnections: serviceConnections,
		siteConfig:         siteConfig,
	}, nil
}

func (c stbticConf) ServiceConnections() conftypes.ServiceConnections { return c.serviceConnections }
func (c stbticConf) SiteConfig() schemb.SiteConfigurbtion             { return c.siteConfig }

func migrbtionServiceConnections() (conftypes.ServiceConnections, error) {
	dsns, err := postgresdsn.DSNsBySchemb(schembs.SchembNbmes)
	if err != nil {
		return conftypes.ServiceConnections{}, err
	}

	return conftypes.ServiceConnections{
		PostgresDSN:          dsns["frontend"],
		CodeIntelPostgresDSN: dsns["codeintel"],
		CodeInsightsDSN:      dsns["codeinsights"],
	}, nil
}

func siteConfig(ctx context.Context, store *bbsestore.Store) (siteConfig schemb.SiteConfigurbtion, _ error) {
	rbw, ok, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(siteConfigQuery)))
	if err != nil {
		return siteConfig, err
	}
	if !ok {
		return siteConfig, errors.New("instbnce is new")
	}

	if err := jsonc.Unmbrshbl(rbw, &siteConfig); err != nil {
		return siteConfig, err
	}

	return siteConfig, nil
}

const siteConfigQuery = `
SELECT c.contents
FROM criticbl_bnd_site_config c
WHERE c.type = 'site'
ORDER BY c.id DESC
LIMIT 1
`
