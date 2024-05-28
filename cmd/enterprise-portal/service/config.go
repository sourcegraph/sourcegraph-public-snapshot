package service

import (
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// Config is the configuration for the Enterprise Portal.
type Config struct {
	DotComDB struct {
		cloudsql.ConnConfig

		PGDSNOverride *string
	}

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

	c.SAMS.ConnConfig = sams.NewConnConfigFromEnv(env)
	c.SAMS.ClientID = env.Get("ENTERPRISE_PORTAL_SAMS_CLIENT_ID", "",
		"Sourcegraph Accounts Management System client ID")
	c.SAMS.ClientSecret = env.Get("ENTERPRISE_PORTAL_SAMS_CLIENT_SECRET", "",
		"Sourcegraph Accounts Management System client secret")
}
