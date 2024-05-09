package cloudsql

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ConnConfig struct {
	// ConnectionName is the CloudSQL connection name,
	// e.g. '${project}:${region}:${instance}'
	ConnectionName *string
	// User is the Cloud SQL user to connect as, e.g. 'test-sa@test-project.iam'
	User *string
	// Database to connect to.
	Database string
	// DialOptions are any additional options to pass to the underlying
	// cloud-sql-proxy driver.
	DialOptions []cloudsqlconn.DialOption
}

// Open opens a *sql.DB connection to the Cloud SQL instance specified by the
// ConnConfig.
//
// 🔔 If you are connecting to a MSP-provisioned Cloud SQL instance,
// DO NOT use this - instead, use runtime.Contract.PostgreSQL.OpenDatabase
// instead.
func Open(
	ctx context.Context,
	cfg ConnConfig,
) (*sql.DB, error) {
	config, err := getCloudSQLConnConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "get CloudSQL connection config")
	}
	return sql.Open("pgx", stdlib.RegisterConnConfig(config))
}

// Connect opens a *pgx.Conn connection to the CloudSQL instance specified by
// the ConnConfig.
//
// 🔔 If you are connecting to a MSP-provisioned Cloud SQL instance,
// DO NOT use this - instead, use runtime.Contract.PostgreSQL.OpenDatabase
// instead.
func Connect(
	ctx context.Context,
	cfg ConnConfig,
) (*pgx.Conn, error) {
	config, err := getCloudSQLConnConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "get CloudSQL connection config")
	}
	return pgx.ConnectConfig(ctx, config)
}

// getCloudSQLConnConfig generates a pgx connection configuration for using
// a Cloud SQL instance using IAM auth.
func getCloudSQLConnConfig(
	ctx context.Context,
	cfg ConnConfig,
) (*pgx.ConnConfig, error) {
	if cfg.ConnectionName == nil || cfg.User == nil {
		return nil, errors.New("missing required PostgreSQL configuration")
	}

	// https://github.com/GoogleCloudPlatform/cloud-sql-go-connector?tab=readme-ov-file#automatic-iam-database-authentication
	dsn := fmt.Sprintf("user=%s dbname=%s", *cfg.User, cfg.Database)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "pgx.ParseConfig")
	}
	customDialer, err := cloudsqlconn.NewDialer(ctx,
		// always the case when using Cloud SQL in MSP
		cloudsqlconn.WithIAMAuthN(),
		// allow passthrough of additional dial options
		cloudsqlconn.WithDefaultDialOptions(cfg.DialOptions...))
	if err != nil {
		return nil, errors.Wrap(err, "cloudsqlconn.NewDialer")
	}
	// Use the Cloud SQL connector to handle connecting to the instance.
	// This approach does *NOT* require the Cloud SQL proxy.
	config.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return customDialer.Dial(ctx, *cfg.ConnectionName)
	}
	return config, nil
}
