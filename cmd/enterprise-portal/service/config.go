package service

import (
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Config is the configuration for the Enterprise Portal.
type Config struct {
	DotComDB struct {
		cloudsql.ConnConfig

		PGDSNOverride *string

		IncludeProductionLicenses bool
	}

	// If nil, no connection was configured.
	CodyGatewayEvents *codygatewayevents.ServiceBigQueryOptions

	SAMS SAMSConfig
}

type SAMSConfig struct {
	sams.ConnConfig
	ClientID     string
	ClientSecret string
}

func (c *Config) Load(env *runtime.Env) {
	c.DotComDB.ConnConfig = cloudsql.ConnConfig{
		ConnectionName: env.GetOptional("DOTCOM_CLOUDSQL_CONNECTION_NAME",
			"Sourcegraph.com Cloud SQL connection name"),
		User:     env.GetOptional("DOTCOM_CLOUDSQL_USER", "Sourcegraph.com Cloud SQL user"),
		Database: env.Get("DOTCOM_CLOUDSQL_DATABASE", "sourcegraph", "Sourcegraph.com database"),
	}
	c.DotComDB.PGDSNOverride = env.GetOptional("DOTCOM_PGDSN_OVERRIDE",
		"For local dev: custom PostgreSQL DSN, overrides DOTCOM_CLOUDSQL_* options")
	c.DotComDB.IncludeProductionLicenses = env.GetBool("DOTCOM_INCLUDE_PRODUCTION_LICENSES", "false",
		"Include production licenses in API results")

	c.SAMS.ConnConfig = sams.NewConnConfigFromEnv(env)
	c.SAMS.ClientID = env.Get("ENTERPRISE_PORTAL_SAMS_CLIENT_ID", "",
		"Sourcegraph Accounts Management System client ID")
	c.SAMS.ClientSecret = env.Get("ENTERPRISE_PORTAL_SAMS_CLIENT_SECRET", "",
		"Sourcegraph Accounts Management System client secret")

	codyGatewayEventsProjectID := env.GetOptional("CODY_GATEWAY_EVENTS_PROJECT_ID",
		"Project ID for Cody Gateway events ('telligentsourcegraph' or 'cody-gateway-dev')")
	codyGatewayEventsDataset := env.Get("CODY_GATEWAY_EVENTS_DATASET", "cody_gateway",
		"Dataset for Cody Gateway events")
	codyGatewayEventsTable := env.Get("CODY_GATEWAY_EVENTS_TABLE", "events",
		"Table for Cody Gateway events")
	if codyGatewayEventsProjectID != nil {
		c.CodyGatewayEvents = &codygatewayevents.ServiceBigQueryOptions{
			ProjectID:   pointers.DerefZero(codyGatewayEventsProjectID),
			Dataset:     codyGatewayEventsDataset,
			EventsTable: codyGatewayEventsTable,
		}
	}
}
