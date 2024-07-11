package contract

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"text/template"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/cloudsql"
)

type postgreSQLContract struct {
	customDSNTemplate *string
	logQueries        bool

	instanceConnectionName *string
	instanceConnectionUser *string
}

func loadPostgreSQLContract(env *Env, isMSP bool) postgreSQLContract {
	c := postgreSQLContract{
		customDSNTemplate: env.GetOptional("PGDSN",
			"Local dev only: custom PostgreSQL DSN with templatized database, e.g. 'user=foo database={{ .Database }}'"),
		logQueries: env.GetBool("PG_QUERY_LOGGING", "false",
			"Local dev only: dump all queries to log output"),

		instanceConnectionName: env.GetOptional("PGINSTANCE", "Cloud SQL instance connection name"),
		instanceConnectionUser: env.GetOptional("PGUSER", "Cloud SQL user"),
	}

	if isMSP && c.customDSNTemplate != nil {
		env.AddError(errors.New("PGDSN is not allowed with MSP=true"))
	}
	if isMSP && c.logQueries {
		env.AddError(errors.New("PG_QUERY_LOGGING=true is not allowed with MSP=true"))
	}

	return c
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
		config, err := parseCustomDSNTemplateConnConfig(*c.customDSNTemplate, database, c.logQueries)
		if err != nil {
			return nil, err
		}
		return sql.Open("pgx", stdlib.RegisterConnConfig(config.ConnConfig))
	}
	return cloudsql.Open(ctx, c.getCloudSQLConnConfig(database))
}

// GetConnectionPool is an alternative to OpenDatabase that returns a
// github.com/jackc/pgx/v5/pgxpool for connecting to the configured database
// instead, for services that prefer to use 'pgx' directly. A pool returns
// without waiting for any connections to be established. Acquire a connection
// immediately after creating the pool to check if a connection can successfully
// be established.
//
// In development, the connection can be overridden with the PGDSN environment
// variable.
func (c postgreSQLContract) GetConnectionPool(ctx context.Context, database string) (*pgxpool.Pool, error) {
	if c.customDSNTemplate != nil {
		config, err := parseCustomDSNTemplateConnConfig(*c.customDSNTemplate, database, c.logQueries)
		if err != nil {
			return nil, err
		}
		return pgxpool.NewWithConfig(ctx, config)
	}
	return cloudsql.GetConnectionPool(ctx, c.getCloudSQLConnConfig(database))
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

func parseCustomDSNTemplateConnConfig(customDSNTemplate, database string, logQueries bool) (*pgxpool.Config, error) {
	tmpl, err := template.New("PGDSN").Parse(customDSNTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "PGDSN is not a valid template")
	}
	var dsn bytes.Buffer
	if err := tmpl.Execute(&dsn, struct{ Database string }{Database: database}); err != nil {
		return nil, errors.Wrap(err, "PGDSN template is invalid")
	}
	config, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, errors.Wrap(err, "rendered PGDSN is invalid")
	}

	if logQueries {
		logger := log.Scoped(fmt.Sprintf("pgx.devtracer.%s", database))
		logger.Warn("enabling query logging")
		config.ConnConfig.Tracer = pgxLocalDevTracer{
			// Just use new root-scoped logger here, as this is local dev, so we
			// don't worry too much about scope propagation.
			logger: logger,
		}
	}

	return config, nil
}

// pgxLocalDevTracer implements various pgx tracing hooks for dumping diagnostics
// in local dev. DO NOT USE OUTSIDE LOCAL DEV.
type pgxLocalDevTracer struct {
	logger log.Logger
}

// Select tracing hooks we want to implement.
var (
	_ pgx.QueryTracer   = pgxLocalDevTracer{}
	_ pgx.ConnectTracer = pgxLocalDevTracer{}
	// Future:
	// _ pgx.BatchTracer    = pgxTracer{}
	// _ pgx.CopyFromTracer = pgxTracer{}
	// _ pgx.PrepareTracer  = pgxTracer{}
)

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls. The returned context is used for the
// rest of the call and will be passed to TraceQueryEnd.
func (t pgxLocalDevTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	var args []string
	for _, arg := range data.Args {
		data, _ := json.Marshal(arg)
		args = append(args, string(data))
	}
	t.logger.Debug(fmt.Sprintf("pgx.QueryStart\n---\n%s\n---\n", data.SQL),
		log.Strings("args", args))
	return ctx
}

func (t pgxLocalDevTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.logger.Debug("pgx.QueryEnd", log.String("commandTag",
		data.CommandTag.String()),
		log.Error(data.Err))
}

func (t pgxLocalDevTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	t.logger.Debug("pgx.ConnectStart",
		log.String("database", data.ConnConfig.Database),
		log.String("instance", fmt.Sprintf("%s:%d", data.ConnConfig.Host, data.ConnConfig.Port)),
		log.String("user", data.ConnConfig.User))
	return ctx
}

func (t pgxLocalDevTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	t.logger.Debug("pgx.ConnectEnd", log.Error(data.Err))
}
