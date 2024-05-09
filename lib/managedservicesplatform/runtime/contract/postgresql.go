package contract

import (
	"bytes"
	"context"
	"database/sql"
	"text/template"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
)

type postgreSQLContract struct {
	customDSNTemplate *string

	instanceConnectionName *string
	instanceConnectionUser *string
}

func loadPostgreSQLContract(env *Env) postgreSQLContract {
	return postgreSQLContract{
		customDSNTemplate: env.GetOptional("PGDSN",
			"custom PostgreSQL DSN with templatized database, e.g. 'user=foo database={{ .Database }}'"),

		instanceConnectionName: env.GetOptional("PGINSTANCE", "Cloud SQL instance connection name"),
		instanceConnectionUser: env.GetOptional("PGUSER", "Cloud SQL user"),
	}
}

// Configured indicates if a PostgreSQL instance is configured for use. It does
// not guarantee the presence of any databases within the instance.
func (c postgreSQLContract) Configured() bool {
	return c.customDSNTemplate != nil ||
		(c.instanceConnectionName != nil && c.instanceConnectionUser == nil)
}

// OpenDatabase returns a standard library DB pointing to the configured
// PostgreSQL database. In MSP, we connect to a Cloud SQL instance over IAM auth.
//
// In development, the connection can be overridden with the PGDSN environment
// variable.
func (c postgreSQLContract) OpenDatabase(ctx context.Context, database string) (*sql.DB, error) {
	if c.customDSNTemplate != nil {
		config, err := parseCustomDSNTemplateConnConfig(*c.customDSNTemplate, database)
		if err != nil {
			return nil, err
		}
		return sql.Open("customdsn", stdlib.RegisterConnConfig(config))
	}
	return cloudsql.Open(ctx, c.getCloudSQLConnConfig(database))
}

// ConnectToDatabase is similar to OpenDatabase, but returns a
// github.com/jackc/pgx/v5 connection to the configured datbase instead for
// services that prefer to use 'pgx' directly.
//
// In development, the connection can be overridden with the PGDSN environment
// variable.
func (c postgreSQLContract) ConnectToDatabase(ctx context.Context, database string) (*pgx.Conn, error) {
	if c.customDSNTemplate != nil {
		config, err := parseCustomDSNTemplateConnConfig(*c.customDSNTemplate, database)
		if err != nil {
			return nil, err
		}
		return pgx.ConnectConfig(ctx, config)
	}
	return cloudsql.Connect(ctx, c.getCloudSQLConnConfig(database))
}

func (c postgreSQLContract) getCloudSQLConnConfig(database string) cloudsql.ConnConfig {
	return cloudsql.ConnConfig{
		ConnectionName: c.instanceConnectionName,
		User:           c.instanceConnectionUser,
		Database:       database,
		DialOptions: []cloudsqlconn.DialOption{
			// MSP-provisioned databases only allow private IP access
			cloudsqlconn.WithPrivateIP(),
		},
	}
}

func parseCustomDSNTemplateConnConfig(customDSNTemplate, database string) (*pgx.ConnConfig, error) {
	tmpl, err := template.New("PGDSN").Parse(customDSNTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "PGDSN is not a valid template")
	}
	var dsn bytes.Buffer
	if err := tmpl.Execute(&dsn, struct{ Database string }{Database: database}); err != nil {
		return nil, errors.Wrap(err, "PGDSN template is invalid")
	}
	config, err := pgx.ParseConfig(dsn.String())
	if err != nil {
		return nil, errors.Wrap(err, "rendered PGDSN is invalid")
	}
	return config, nil
}
