package shared

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type staticConf struct {
	serviceConnections conftypes.ServiceConnections
	siteConfig         schema.SiteConfiguration
}

func newStaticConf(ctx context.Context, db database.DB) (*staticConf, error) {
	serviceConnections, err := serviceConnections()
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

func serviceConnections() (conftypes.ServiceConnections, error) {
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
